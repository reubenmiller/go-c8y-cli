package c8ylogin

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/cli/browser"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/mdp/qrterminal/v3"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/prompt"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/pkg/oauth/api"
	"github.com/reubenmiller/go-c8y/pkg/oauth/device"
)

// LoginState current state of the login flow
type LoginState int

func (l LoginState) String() string {
	switch l {
	case LoginStateAuth:
		return "Authorized"
	case LoginStateNoAuth:
		return "NotAuthorized"
	case LoginStateTFASetup:
		return "TFASetup"
	case LoginStateTFAConfirm:
		return "TFAConfirm"
	case LoginStateLogin:
		return "Login"
	case LoginStateVerify:
		return "Verify"
	case LoginStateAbort:
		return "Abort"
	case LoginStatePromptPassword:
		return "PromptForPassword"
	case LoginStateUnknown:
		fallthrough
	default:
		return "Unknown"
	}
}

const (
	// LoginStateUnknown unknown
	LoginStateUnknown LoginState = iota

	// LoginStateAuth client is authorized
	LoginStateAuth

	// LoginStateNoAuth missing client authorization
	LoginStateNoAuth

	// LoginStateTFASetup user requires TFA setup
	LoginStateTFASetup

	// LoginStateTFAConfirm user requires TFA setup to be confirmed
	LoginStateTFAConfirm

	// LoginStateLogin login (mostly used for OAUTH2 / SSO)
	LoginStateLogin

	// LoginStateVerify verify authorization state by sending a request to Cumulocity
	LoginStateVerify

	// LoginStateAbort abort the login flow
	LoginStateAbort

	// LoginStatePromptPassword prompt for the password
	LoginStatePromptPassword
)

var (
	// Timeout in seconds used for all Server requests
	Timeout int = 30000
)

// LoginHandler handler to process all login / authorization tasks for Cumulocity.
// Two-Factor authentication, TFA setup, OAUTH etc.
type LoginHandler struct {
	IO              *iostreams.IOStreams
	TFACodeRequired bool
	Authorized      bool
	Err             error
	TFACode         string
	C8Yclient       *c8y.Client
	LoginOptions    *c8y.TenantLoginOptions
	state           chan LoginState
	Attempts        int
	Writer          io.Writer
	Logger          *logger.Logger
	LoginType       string
	LoginAttempted  bool

	// SSO specific settings
	SSO config.SSOSettings

	onSave func()
}

// NewLoginHandler creates a new login handler to process the full Cumulocity login process for different login types, i.e. OAUTH2_INTERNAL, BASIC etc.
func NewLoginHandler(IO *iostreams.IOStreams, c *c8y.Client, w io.Writer, onSave func()) *LoginHandler {
	h := &LoginHandler{
		IO:        IO,
		C8Yclient: c,
		Writer:    w,
		onSave:    onSave,
		Logger:    logger.NewDummyLogger("c8ylogin"),
		SSO:       config.SSOSettings{},
	}
	h.state = make(chan LoginState, 1)
	return h
}

// SetLogger sets the logger to use to
func (lh *LoginHandler) SetLogger(l *logger.Logger) {
	lh.Logger = l
}

func (lh *LoginHandler) do(op func() error) {
	lh.Err = op()
}

// Clear clears any existing authorization state stored for the current client
func (lh *LoginHandler) Clear() {
	lh.Logger.Info("Clearing authentication")

	if lh.Attempts > 1 && lh.TFACodeRequired {
		lh.Logger.Infof("Clearing TFA code. code=%s", lh.TFACode)
		lh.TFACode = ""
	}
	lh.Authorized = false
	// lh.TFACodeRequired = false
	lh.C8Yclient.ClearToken()
	lh.onSave()
	lh.C8Yclient.SetAuthorizationType(c8y.AuthTypeBasic)
	lh.Err = nil
}

// Run initiates the authorization process. It will be trigger the whole login flow and prompt the user for any information that is no
// available otherwise.
func (lh *LoginHandler) Run() error {
	lh.init()

	// Check if any authentication is set
	if lh.LoginType == c8y.LoginTypeNone {
		lh.state <- LoginStateVerify
	} else if lh.LoginType == c8y.LoginTypeOAuth2Internal {
		if lh.C8Yclient.Token != "" {
			lh.state <- LoginStateVerify
		} else {
			lh.state <- LoginStateLogin
		}
	} else if lh.C8Yclient.Token != "" {
		lh.state <- LoginStateVerify
	} else if lh.LoginType == c8y.LoginTypeOAuth2 {
		lh.state <- LoginStateLogin
	} else if lh.C8Yclient.Password != "" {
		lh.state <- LoginStateVerify
	} else if lh.C8Yclient.Username == "" {
		// OAUTH2 doesn't provide a username
		lh.state <- LoginStateLogin
	} else {
		lh.state <- LoginStatePromptPassword
	}

	for {
		c := <-lh.state

		lh.Logger.Infof("Current State: %s", c.String())

		if c == LoginStateUnknown {
			lh.Clear()
			lh.verify()
		} else if c == LoginStateVerify {
			lh.verify()
		} else if c == LoginStateTFAConfirm {
			lh.login()
		} else if c == LoginStateLogin {
			lh.login()
		} else if c == LoginStateNoAuth {
			lh.Clear()

			// login
			lh.login()

		} else if c == LoginStatePromptPassword {
			if err := lh.promptForPassword(); err != nil {
				return err
			}
		} else if c == LoginStateTFASetup {
			if err := lh.setupTFA(); err != nil {
				return fmt.Errorf("TFA Setup failed")
			}
			lh.state <- LoginStateTFAConfirm
		} else if c == LoginStateAbort {
			break
		} else if c == LoginStateAuth {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	//
	// When reusing an existing token, the default login type might be wrong so
	// try and detect the type by parsing the token
	if !lh.LoginAttempted {
		if lh.C8Yclient.Token != "" {
			if subject, err := c8ysession.GetTokenSubject(lh.C8Yclient.Token); err == nil {
				// with c8y issued tokens, the subject is the username
				if lh.C8Yclient.Username == subject {
					lh.LoginType = c8y.LoginTypeOAuth2Internal
				} else {
					lh.LoginType = c8y.LoginTypeOAuth2
				}
			}
		}
	}

	return lh.Err
}

func (lh *LoginHandler) sortLoginOptions() {
	if lh.LoginOptions == nil {
		return
	}

	optionOrder := map[string]int{
		c8y.LoginTypeNone:           40,
		c8y.LoginTypeBasic:          30,
		c8y.LoginTypeOAuth2Internal: 20,
		c8y.LoginTypeOAuth2:         10,
	}

	if lh.LoginType != "" {
		if _, ok := optionOrder[lh.LoginType]; ok {
			lh.Logger.Infof("Setting preferred login method. %s", lh.LoginType)
			optionOrder[lh.LoginType] = -999 // Try preferred method first
		} else {
			lh.Logger.Infof("Unsupported login method. The given option will be ignored. %s", lh.LoginType)
		}
	}

	// sort login options
	sort.SliceStable(lh.LoginOptions.LoginOptions, func(i, j int) bool {
		iWeight := 100
		jWeight := 200

		if v, ok := optionOrder[lh.LoginOptions.LoginOptions[i].Type]; ok {
			iWeight = v
		}

		if v, ok := optionOrder[lh.LoginOptions.LoginOptions[j].Type]; ok {
			jWeight = v
		}
		return iWeight < jWeight
	})
}

func (lh *LoginHandler) init() {
	lh.do(func() error {
		// Special case where no auth is required
		if lh.LoginType == c8y.LoginTypeNone {
			return nil
		}
		loginOptions, _, err := lh.C8Yclient.Tenant.GetLoginOptions(context.Background())

		if err != nil {
			if strings.Contains(err.Error(), "401") {
				lh.Logger.Infof("Login options returned 401. This may happen when your c8y hostname is not setup correctly")
				return nil
			}
			lh.Logger.Warnf("Failed to get login options. %s", err)
			return err
		}
		if loginOptions == nil {
			lh.Logger.Warnf("Login options are empty. %s", err)
			return err
		}

		if lh.C8Yclient.TenantName != "" {
			tenantSelfURL := ""

			if len(loginOptions.LoginOptions) > 0 {
				tenantSelfURL = loginOptions.LoginOptions[0].Self
			} else {
				tenantSelfURL = loginOptions.Self
			}
			if strings.HasPrefix(tenantSelfURL, "http") && !strings.Contains(tenantSelfURL, lh.C8Yclient.TenantName+".") {
				lh.Logger.Warningf("Detected invalid tenant name. expected %s to include %s. Tenant name will be ignored", lh.C8Yclient.TenantName, tenantSelfURL)
				lh.C8Yclient.TenantName = ""
			}
		}

		lh.LoginOptions = loginOptions
		lh.sortLoginOptions()

		if len(lh.LoginOptions.LoginOptions) > 0 {
			lh.LoginType = lh.LoginOptions.LoginOptions[0].Type

			// Setting preferred login method
			lh.Logger.Debugf("Preferred login method. type=%s", lh.LoginType)
		}
		return err
	})
}

func emptyPointer(ignored []rune) []rune {
	return []rune("")
}

func (lh *LoginHandler) promptForPassword() error {
	validate := func(input string) error {
		if input == "" {
			return fmt.Errorf("Empty password")
		}
		return nil
	}

	if lh.C8Yclient.Username == "" {
		prompter := prompt.NewPrompt(lh.Logger)
		username, err := prompter.Username("Enter username", " ")
		if err != nil {
			lh.state <- LoginStateAbort
			lh.Err = fmt.Errorf("user cancelled prompt")
			return lh.Err
		}
		lh.C8Yclient.Username = username
	}

	reason := ""
	label := fmt.Sprintf("Enter c8y password for user (%s)", lh.C8Yclient.Username)

	// Provide additional information to user what happened
	// invalid encryption key? or missing password
	if lh.C8Yclient.Password == "" {
		reason = "password is empty"
	} else {
		reason = "password is invalid"
		label = fmt.Sprintf("Re-enter c8y password for user (%s)", lh.C8Yclient.Username)
	}

	prompt := promptui.Prompt{
		Stdin:       os.Stdin,
		Stdout:      os.Stderr,
		Default:     "",
		Mask:        ' ',
		HideEntered: true,
		Pointer:     emptyPointer,
		Label:       fmt.Sprintf("%s 🔒", label),
		Validate:    validate,
		Templates:   &promptui.PromptTemplates{},
	}
	if reason != "" {
		prompt.Templates.Invalid = fmt.Sprintf("{{ \"✗ (%s)\" | red | faint }} {{ . | bold }}: {{ \"[input is hidden]\" | faint }}", reason)
		prompt.Templates.Valid = fmt.Sprintf("{{ \"✗ (%s)\" | red | faint }} {{ . | bold }}: {{ \"[input is hidden]\" | faint }}", reason)
	}
	pass, err := prompt.Run()

	if err != nil {
		return err
	}
	lh.C8Yclient.Password = pass
	lh.state <- LoginStateVerify
	return nil
}
func (lh *LoginHandler) login() {
	lh.do(func() error {
		if lh.LoginOptions == nil || len(lh.LoginOptions.LoginOptions) == 0 {
			// Don't fail on no login options. Default to using BASIC auth
			lh.Logger.Infof("No login options available. Using BASIC auth")
			if lh.Authorized {
				lh.state <- LoginStateAuth
			} else {
				lh.LoginType = c8y.LoginTypeBasic
				lh.state <- LoginStateVerify
			}
			return nil
		}

		if lh.Authorized {
			lh.state <- LoginStateAuth
			return nil
		}

		lh.LoginAttempted = true

		order := []string{}
		for _, option := range lh.LoginOptions.LoginOptions {
			order = append(order, option.Type)
		}
		lh.Logger.Infof("Login type order: %v", order)

		for _, option := range lh.LoginOptions.LoginOptions {
			lh.Logger.Infof("Trying login option. %s", option.Type)
			switch option.Type {
			case c8y.LoginTypeOAuth2:
				// Device Authorization Flow (for external providers)
				if !lh.IO.CanPromptOnStdErr() {
					lh.state <- LoginStateAbort
					lh.Err = fmt.Errorf("OAuth2 device flow requires an interactive console")
					return lh.Err
				}

				displayDeviceCode := func(code *device.CodeResponse) error {
					verificationURI := code.VerificationURIComplete

					lh.writeMessage("\nLogging in using the device flow (OAUTH2)\n\n")

					bold := color.New(color.Bold)
					bold.EnableColor()

					if verificationURI == "" {
						verificationURI = code.VerificationURI
						fmt.Fprintf(os.Stderr, " First copy your one-time code: %s\n", bold.Sprint(code.UserCode))
					}

					fmt.Fprintf(lh.IO.ErrOut, "%s to open %s in your browser...", bold.Sprintf("Press Enter"), verificationURI)
					bufio.NewReader(lh.IO.In).ReadBytes('\n')

					if browserErr := browser.OpenURL(verificationURI); browserErr != nil {
						lh.writeMessage("Failed to open browser, please open the URL in your browser manually")
					}
					return nil
				}

				accessToken, loginErr := lh.C8Yclient.Tenant.AuthorizeWithDeviceFlow(context.Background(), option.InitRequest, api.AuthEndpoints{
					// Allow users to provide their own discovery URL and scopes
					OpenIDConfigurationURL: lh.SSO.DiscoveryURL,
					Scopes:                 lh.SSO.Scopes,
					AuthRequestOptions: []api.AuthRequestEditorFn{
						api.WithAudience(lh.SSO.Audience),
					},
				}, displayDeviceCode)
				if loginErr != nil {
					// Ignore SSO if invalid configuration is found
					if errors.Is(loginErr, c8y.ErrSSOInvalidConfiguration) {
						lh.Logger.Warnf("Skipping login type (%s) as SSO configuration is invalid. err=%s", option.Type, loginErr)
						continue
					}

					// Check error type, and if the configuration can't be found, then skip SSO
					// add a new formal type to cover this scenario
					// could not get OpenID Connect configuration
					lh.state <- LoginStateAbort
					lh.Err = fmt.Errorf("OAuth2 device authorization flow failed. %w", loginErr)
					return lh.Err
				}

				lh.Logger.Infof("Received access token via device flow. type=%s, scope=%s, tokenPresent=%v, refreshTokenPresent=%v", accessToken.Type, accessToken.Scope, accessToken.Token != "", accessToken.RefreshToken != "")
				lh.onSave()
				lh.state <- LoginStateVerify
				return nil
			case c8y.LoginTypeOAuth2Internal:

				if lh.TFACodeRequired && option.TFAStrategy == "TOTP" {
					os.Stderr.WriteString(fmt.Sprintf("Session details:\nHost=%s, username=%s\n", lh.C8Yclient.BaseURL.Host, lh.C8Yclient.Username))
					if lh.TFACode == "" {
						prompt := promptui.Prompt{
							Stdin:       os.Stdin,
							Stdout:      os.Stderr,
							Default:     lh.TFACode,
							HideEntered: true,
							Label:       "Enter Two-Factor code",
							Validate: func(input string) error {
								if len(strings.ReplaceAll(input, " ", "")) < 6 {
									return fmt.Errorf("Missing TFA code")
								}

								ctx, cancel := context.WithTimeout(context.Background(), time.Duration(Timeout)*time.Millisecond)
								defer cancel()

								lh.C8Yclient.TFACode = input

								if err := lh.C8Yclient.LoginUsingOAuth2(ctx, option.InitRequest); err != nil {
									return err
								}
								lh.TFACode = input
								lh.Authorized = true
								return nil
							},
						}

						if v, err := prompt.Run(); err == nil {
							lh.TFACode = v
						} else {
							lh.state <- LoginStateAbort
							lh.Err = fmt.Errorf("User cancelled login")
							return lh.Err
						}
					}
					lh.C8Yclient.TFACode = lh.TFACode
				}

				if !lh.Authorized {
					ctx, cancel := context.WithTimeout(context.Background(), time.Duration(Timeout)*time.Millisecond)
					defer cancel()

					lh.Logger.Debugf("Logging in using %s", c8y.LoginTypeOAuth2Internal)

					// Check if username/password are provided
					if lh.C8Yclient.Username == "" {
						lh.Logger.Warnf("Skipping login type (%s) as a username is empty", option.Type)
						continue
					}
					if lh.C8Yclient.Password == "" {
						lh.Logger.Warnf("Skipping login type (%s) as password is empty", option.Type)
						continue
					}

					if err := lh.C8Yclient.LoginUsingOAuth2(ctx, option.InitRequest); err != nil {
						if v, ok := err.(*c8y.ErrorResponse); ok {
							lh.Logger.Errorf("OAuth2 failed. %s", v.Message)
						} else {
							lh.Logger.Errorf("OAuth2 failed. %s", err)
						}

						if strings.Contains(err.Error(), "There was a change in authentication strategy for your tenant or user account") {
							// trigger unknown to recheck if TFA is required or not
							lh.state <- LoginStateUnknown
							return nil
						}

						// trigger unknown to recheck if TFA is required or not
						lh.state <- LoginStateUnknown
						return nil
					}
					lh.Authorized = true
					lh.state <- LoginStateAuth
					lh.onSave()
					return nil
				}
				lh.onSave()
				lh.state <- LoginStateVerify

			case c8y.LoginTypeBasic:
				if lh.C8Yclient.Username == "" || lh.C8Yclient.Password == "" {
					if lh.IO.CanPromptOnStdErr() {
						lh.state <- LoginStatePromptPassword
						return nil
					} else {
						lh.Err = fmt.Errorf("username or password is empty and the interactive prompt is disabled")
						lh.state <- LoginStateAbort
						return lh.Err
					}
				}
				lh.state <- LoginStateVerify
				return nil
			}
		}

		// return an error by default
		lh.state <- LoginStateAbort
		lh.Err = fmt.Errorf("no valid login type found")
		return lh.Err
	})
}

func (lh *LoginHandler) errorContains(message, pattern string) bool {
	return strings.Contains(strings.ToLower(message), strings.ToLower(pattern))
}

func (lh *LoginHandler) verify() {
	lh.do(func() error {
		lh.Attempts++

		if lh.Attempts > 2 {
			lh.Err = fmt.Errorf("max log attempts reached")
			lh.state <- LoginStateAbort
			return lh.Err
		}

		tenant, resp, err := lh.C8Yclient.Tenant.GetCurrentTenant(context.Background())

		if resp != nil && resp.StatusCode() == http.StatusUnauthorized {

			if v, ok := err.(*c8y.ErrorResponse); ok {
				lh.Logger.Infof("error message from server. %s", v.Message)

				if lh.errorContains(v.Message, "TFA TOTP setup required") {
					lh.TFACodeRequired = true
					lh.state <- LoginStateTFASetup
				} else if lh.errorContains(v.Message, "TFA TOTP code required") {
					lh.Logger.Debugf("TFA code is required. server response: %s", v.Message)
					lh.TFACodeRequired = true
					lh.state <- LoginStateNoAuth
				} else if lh.errorContains(v.Message, "User has been logged out") {
					lh.Logger.Warning("User had been logged out. Clearing token and trying again")
					lh.state <- LoginStateNoAuth
					lh.C8Yclient.ClearToken()
					lh.onSave()
				} else if lh.errorContains(v.Message, "Tenant has no access from outside the platform") {
					// Decide what to do here
					lh.LoginType = c8y.LoginTypeBasic
					lh.state <- LoginStateVerify
				} else if lh.errorContains(v.Message, "Bad credentials") || lh.errorContains(v.Message, "Invalid credentials") {
					lh.Logger.Infof("Bad credentials, using auth method: %s", lh.C8Yclient.AuthorizationType)

					// try resetting the tenant (in case if it is incorrect)
					lh.C8Yclient.TenantName = ""

					if lh.C8Yclient.AuthorizationType != c8y.AuthTypeBearer {
						if lh.IO.CanPromptOnStdErr() {
							lh.state <- LoginStatePromptPassword
						} else {
							lh.state <- LoginStateAbort
						}
					} else {
						lh.state <- LoginStateNoAuth
					}
				} else {
					lh.state <- LoginStateNoAuth
				}
				return nil
			}

			lh.state <- LoginStateNoAuth
		} else {
			lh.state <- LoginStateAuth
			lh.Logger.Infof("Detected tenant: %s", tenant.Name)
			lh.C8Yclient.TenantName = tenant.Name

			// Get Cumulocity system version
			if lh.C8Yclient.Version == "" {
				if version, err := lh.C8Yclient.TenantOptions.GetVersion(context.Background()); err == nil {
					lh.C8Yclient.Version = version
				} else {
					lh.Logger.Warnf("Could not get Cumulocity System version. %s", err)
				}
			}

			// Get user information if not set (e.g. external SSO tokens don't generally include the username in a known field)
			if lh.C8Yclient.Username == "" {
				lh.Logger.Infof("Getting the current user's information as the username is not set")
				currentUser, _, err := lh.C8Yclient.User.GetCurrentUser(context.Background())
				if err != nil {
					lh.Logger.Warnf("Could not get username. %s", err)
				} else {
					lh.Logger.Infof("Found current username. %s", currentUser.Username)
					lh.C8Yclient.Username = currentUser.Username
				}
			}
		}
		return err
	})
}

func (lh LoginHandler) writeMessage(m string) {
	if _, err := io.WriteString(lh.Writer, m); err != nil {
		lh.Logger.Warnf("Failed to write message. %s", err)
	}
}

func (lh LoginHandler) writeMessageF(format string, a interface{}) {
	if _, err := io.WriteString(lh.Writer, fmt.Sprintf(format, a)); err != nil {
		lh.Logger.Warnf("Failed to write message. %s", err)
	}
}

func (lh *LoginHandler) setupTFA() error {

	// Request TFA secret
	authFunc := c8y.WithTenantUsernamePassword(
		lh.C8Yclient.TenantName,
		lh.C8Yclient.Username,
		lh.C8Yclient.Password,
	)
	resp, err := lh.C8Yclient.SendRequest(
		context.Background(),
		c8y.RequestOptions{
			Method:   http.MethodPost,
			Path:     "/user/currentUser/totpSecret",
			AuthFunc: authFunc,
		})

	if err != nil {
		lh.Logger.Infof("Could not get tot")
		return err
	}

	lh.Logger.Infof("secret-raw: %s", resp.String())

	secretQRURL := ""
	if v := resp.JSON("secretQrUrl"); v.Exists() {
		secretQRURL = v.String()
	}

	// Display TOTP secret
	if v := resp.JSON("rawSecret"); v.Exists() {
		totpURL := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s", lh.C8Yclient.Username, v.String(), lh.C8Yclient.BaseURL.Host)
		qrterminal.GenerateWithConfig(totpURL, qrterminal.Config{
			Level:      qrterminal.M,
			Writer:     lh.Writer,
			HalfBlocks: true,
			BlackChar:  qrterminal.BLACK,
			WhiteChar:  qrterminal.WHITE,
			QuietZone:  1,
		})

		lh.writeMessage("\n")
		if secretQRURL != "" {
			lh.writeMessageF("TOTP URL: %s\n", secretQRURL)
		}

		lh.writeMessageF("TOTP Secret: %s\n\n", v.String())
	}

	// Verify TOTP by checking a code
	tfaCodePrompt := promptui.Prompt{
		Stdin:       os.Stdin,
		Stdout:      os.Stderr,
		HideEntered: true,
		Label:       "Enter Two-Factor code",
		Validate:    lh.verifyTFASetupCode,
	}

	if _, err := tfaCodePrompt.Run(); err != nil {
		return err
	}

	// Activate totp
	_, err = lh.C8Yclient.SendRequest(
		context.Background(),
		c8y.RequestOptions{
			Method:   http.MethodPost,
			Path:     "/user/currentUser/totpSecret/activity",
			Body:     map[string]interface{}{"isActive": true},
			AuthFunc: authFunc,
		},
	)

	if err != nil {
		return fmt.Errorf("Failed to activate TFA (TOTP): %w", err)
	}

	time.Sleep(1000 * time.Millisecond)
	return nil
}

func (lh *LoginHandler) verifyTFASetupCode(input string) error {
	if len(strings.ReplaceAll(input, " ", "")) < 6 {
		return fmt.Errorf("Non-zero input")
	}

	_, err := lh.C8Yclient.SendRequest(
		context.Background(),
		c8y.RequestOptions{
			Method: http.MethodPost,
			Path:   "/user/currentUser/totpSecret/verify",
			Body:   map[string]interface{}{"code": input},
		},
	)

	if err != nil {
		return err
	}
	lh.TFACode = input
	return nil
}
