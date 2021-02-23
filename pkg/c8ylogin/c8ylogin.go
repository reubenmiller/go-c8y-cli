package c8ylogin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mdp/qrterminal"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

// LoginState current state of the login flow
type LoginState int

func (l LoginState) String() string {
	return [...]string{"Unknown", "Authorized", "NotAuthorized", "TFASetup", "TFAConfirm", "Verify", "Abort", "PromptForPassword"}[l]
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
	TFACodeRequired bool
	Authorized      bool
	Interactive     bool
	Err             error
	TFACode         string
	Cookies         []*http.Cookie
	C8Yclient       *c8y.Client
	LoginOptions    *c8y.TenantLoginOptions
	state           chan LoginState
	Attempts        int
	Writer          io.Writer
	Logger          *logger.Logger

	onSave func()
}

// NewLoginHandler creates a new login handler to process the full Cumulocity login process for different login types, i.e. OAUTH_INTERNAL, BASIC etc.
func NewLoginHandler(c *c8y.Client, w io.Writer, onSave func()) *LoginHandler {
	h := &LoginHandler{
		C8Yclient:   c,
		Interactive: true,
		Writer:      w,
		onSave:      onSave,
		Logger:      logger.NewDummyLogger("c8ylogin"),
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
	lh.C8Yclient.SetCookies([]*http.Cookie{})
	lh.onSave()
	lh.C8Yclient.AuthorizationMethod = c8y.AuthMethodBasic
	lh.Err = nil
}

// Run initiates the authorization process. It will be trigger the whole login flow and prompt the user for any information that is no
// available otherwise.
func (lh *LoginHandler) Run() error {
	lh.init()
	lh.state <- LoginStateVerify

	for {
		c := <-lh.state

		lh.Logger.Infof("Current State: %s\n", c.String())

		if c == LoginStateUnknown {
			lh.Clear()
			lh.verify()
		} else if c == LoginStateVerify {
			lh.verify()
		} else if c == LoginStateTFAConfirm {
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

	return lh.Err
}

func (lh *LoginHandler) sortLoginOptions() {
	if lh.LoginOptions == nil {
		return
	}

	optionOrder := map[string]int{
		c8y.AuthMethodBasic:          2,
		c8y.AuthMethodOAuth2Internal: 1,
	}

	// sort login options
	sort.SliceStable(lh.LoginOptions.LoginOptions[:], func(i, j int) bool {
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
		loginOptions, _, err := lh.C8Yclient.Tenant.GetLoginOptions(context.Background())
		if err != nil {
			lh.Logger.Errorf("Failed to get login options. %s", err)
			return err
		}
		lh.LoginOptions = loginOptions
		lh.sortLoginOptions()

		if len(lh.LoginOptions.LoginOptions) > 0 {
			lh.C8Yclient.AuthorizationMethod = lh.LoginOptions.LoginOptions[0].Type

			// Setting preferred login method
			lh.Logger.Debugf("Preferred login method. type=%s", lh.C8Yclient.AuthorizationMethod)
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

	reason := ""
	label := "Enter c8y password"

	// Provide additional information to user what happened
	// invalid encryption key? or missing password
	if lh.C8Yclient.Password == "" {
		reason = "password is empty"
	} else {
		reason = "password is invalid"
		label = "Re-enter c8y password"
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
			lh.state <- LoginStateAbort
			return fmt.Errorf("No login options")
		}

		if lh.Authorized {
			lh.state <- LoginStateAuth
			return nil
		}

		for _, option := range lh.LoginOptions.LoginOptions {
			switch option.Type {
			case c8y.AuthMethodOAuth2Internal:

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
									lh.Logger.Errorf("OAuth2 failed. %s", err)
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
							return nil
						}
					}
					lh.C8Yclient.TFACode = lh.TFACode
				}

				if !lh.Authorized {
					ctx, cancel := context.WithTimeout(context.Background(), time.Duration(Timeout)*time.Millisecond)
					defer cancel()

					lh.Logger.Debugf("Logging in using %s", c8y.AuthMethodOAuth2Internal)
					if err := lh.C8Yclient.LoginUsingOAuth2(ctx, option.InitRequest); err != nil {
						lh.Attempts++

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

						if lh.Attempts > 2 {
							lh.Err = fmt.Errorf("Max log attempts reached: %w", err)
							lh.state <- LoginStateAbort
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

			case c8y.AuthMethodBasic:
				// do nothing
			}
			break
		}
		return nil
	})
}

func (lh *LoginHandler) errorContains(message, pattern string) bool {
	return strings.Contains(strings.ToLower(message), strings.ToLower(pattern))
}

func (lh *LoginHandler) verify() {
	lh.do(func() error {
		// _, resp, err := lh.C8Yclient.User.GetCurrentUser(context.Background())
		tenant, resp, err := lh.C8Yclient.Tenant.GetCurrentTenant(context.Background())

		if resp != nil && resp.StatusCode == http.StatusUnauthorized {

			if v, ok := err.(*c8y.ErrorResponse); ok {
				lh.Logger.Warning("error message from server. %s", v.Message)

				if lh.errorContains(v.Message, "TFA TOTP setup required") {
					lh.TFACodeRequired = true
					lh.state <- LoginStateTFASetup
				} else if lh.errorContains(v.Message, "TFA TOTP code required") {
					lh.Logger.Debug("TFA code is required. server response: %s", v.Message)
					lh.TFACodeRequired = true
					lh.state <- LoginStateNoAuth
				} else if lh.errorContains(v.Message, "User has been logged out") {
					// TODO: Invalidate the cookies
					lh.Logger.Warning("User had been logged out. Clearing cookies and trying again")
					lh.state <- LoginStateNoAuth
					lh.C8Yclient.SetCookies([]*http.Cookie{})
					lh.onSave()
				} else if lh.errorContains(v.Message, "Bad credentials") {
					lh.Logger.Infof("Bad creds using auth method: %s", lh.C8Yclient.AuthorizationMethod)
					if lh.C8Yclient.AuthorizationMethod != c8y.AuthMethodOAuth2Internal {
						if lh.Interactive {
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
		}
		return err
	})
}

func (lh LoginHandler) writeMessage(m string) {
	io.WriteString(lh.Writer, m)
}

func (lh LoginHandler) writeMessageF(format string, a interface{}) {
	io.WriteString(lh.Writer, fmt.Sprintf(format, a))
}

func (lh *LoginHandler) setupTFA() error {

	// Request TFA secret
	backupAuthMethod := lh.C8Yclient.AuthorizationMethod
	lh.C8Yclient.AuthorizationMethod = c8y.AuthMethodBasic
	resp, err := lh.C8Yclient.SendRequest(
		context.Background(),
		c8y.RequestOptions{
			Method: http.MethodPost,
			Path:   "/user/currentUser/totpSecret",
		})

	if err != nil {
		lh.Logger.Infof("Could not get tot")
		return err
	}

	lh.Logger.Infof("secret-raw: %s", *resp.JSONData)

	secretQRURL := ""
	if v := resp.JSON.Get("secretQrUrl"); v.Exists() {
		secretQRURL = v.String()
	}

	// Display TOTP secret
	if v := resp.JSON.Get("rawSecret"); v.Exists() {
		totpURL := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s", lh.C8Yclient.Username, v.String(), lh.C8Yclient.BaseURL.Host)
		qrterminal.GenerateWithConfig(totpURL, qrterminal.Config{
			Level:     qrterminal.M,
			Writer:    lh.Writer,
			BlackChar: qrterminal.BLACK,
			WhiteChar: qrterminal.WHITE,
			QuietZone: 1,
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
	resp, err = lh.C8Yclient.SendRequest(
		context.Background(),
		c8y.RequestOptions{
			Method: http.MethodPost,
			Path:   "/user/currentUser/totpSecret/activity",
			Body:   map[string]interface{}{"isActive": true},
		},
	)

	if err != nil {
		return fmt.Errorf("Failed to activate TFA (TOTP): %w", err)
	}

	time.Sleep(1000 * time.Millisecond)

	lh.C8Yclient.AuthorizationMethod = backupAuthMethod
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
