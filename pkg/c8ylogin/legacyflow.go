package c8ylogin

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logintype"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/prompt"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	authv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/loginoptions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/currenttenant"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/systemoptions"
	"github.com/tidwall/gjson"
)

// This file contains the login flow which is driven by the login options of the tenant, i.e.
// BASIC, OAUTH2_INTERNAL (tokens), and OAUTH2 (external provider using the device flow).
// It is a state machine which verifies the current credentials, and logs in or prompts for the
// missing information when they are not valid.

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

// loginOption a login option of the tenant
type loginOption struct {
	Type        string
	InitRequest string
	TFAStrategy string
	Self        string
}

func emptyPointer(ignored []rune) []rune {
	return []rune("")
}

// errorMessage gets the message of an error from the server
func errorMessage(err error) string {
	var apiErr *core.Error
	if errors.As(err, &apiErr) && apiErr != nil {
		return serverMessage(*apiErr)
	}
	var apiErrValue core.Error
	if errors.As(err, &apiErrValue) {
		return serverMessage(apiErrValue)
	}
	return err.Error()
}

// serverMessage gets the message of the response of the server (without the additional info link)
func serverMessage(err core.Error) string {
	if v := gjson.Get(err.MessageRaw, "message"); v.Exists() {
		return v.String()
	}
	return err.Message
}

func errorContains(message, pattern string) bool {
	return strings.Contains(strings.ToLower(message), strings.ToLower(pattern))
}

func (h *LoginHandlerV2) runTenantFlow() error {
	h.abortErr = nil
	err := h.initLoginOptions()

	state := h.initialState()
	for {
		h.Logger.Infof("Current State: %s", state.String())

		switch state {
		case LoginStateUnknown:
			h.clearAuth()
			err = nil
			state, err = h.verify()
		case LoginStateVerify:
			state, err = h.verify()
		case LoginStateTFAConfirm, LoginStateLogin:
			state, err = h.login()
		case LoginStateNoAuth:
			h.clearAuth()
			state, err = h.login()
		case LoginStatePromptPassword:
			if promptErr := h.promptForPassword(); promptErr != nil {
				return promptErr
			}
			state = LoginStateVerify
		case LoginStateTFASetup:
			// the TFA setup is handled by the login (it is requested by the server when the
			// token is requested, and it needs the credentials of the user)
			state = LoginStateLogin
		}

		if state == LoginStateAbort || state == LoginStateAuth {
			break
		}
	}

	//
	// When reusing an existing token, the default login type might be wrong so
	// try and detect the type by parsing the token
	if !h.loginAttempted {
		if h.auth.Token != "" {
			if subject, subjectErr := c8ysession.GetTokenSubject(h.auth.Token); subjectErr == nil {
				// with c8y issued tokens, the subject is the username
				if h.auth.Username == subject {
					h.LoginType = logintype.OAuth2Internal
				} else {
					h.LoginType = logintype.OAuth2
				}
			}
		}
	}

	if err == nil && h.abortErr != nil {
		err = h.abortErr
	}
	return err
}

func (h *LoginHandlerV2) initialState() LoginState {
	// Check if any authentication is set
	switch {
	case h.LoginType == logintype.None:
		return LoginStateVerify
	case h.LoginType == logintype.OAuth2Internal:
		if h.auth.Token != "" {
			return LoginStateVerify
		}
		return LoginStateLogin
	case h.auth.Token != "":
		return LoginStateVerify
	case h.LoginType == logintype.OAuth2:
		return LoginStateLogin
	case h.auth.Password != "":
		return LoginStateVerify
	case h.auth.Username == "":
		// OAUTH2 doesn't provide a username
		return LoginStateLogin
	default:
		return LoginStatePromptPassword
	}
}

// clearAuth clears any existing authorization state stored for the current client
func (h *LoginHandlerV2) clearAuth() {
	h.Logger.Info("Clearing authentication")

	if h.attempts > 1 && h.tfaCodeRequired {
		h.Logger.Infof("Clearing TFA code. code=%s", h.TOTPCode)
		h.TOTPCode = ""
	}
	h.authorized = false
	h.auth.Token = ""
	h.onSave()
	h.authType = authv2.AuthTypeBasic
	h.syncAuth()
}

// initLoginOptions gets the login options of the tenant, and sets the preferred login type
func (h *LoginHandlerV2) initLoginOptions() error {
	// Special case where no auth is required
	if h.LoginType == logintype.None {
		return nil
	}

	result := h.V2Client.LoginOptions.ListNoAuth(h.ctx(), loginoptions.ListOptions{})
	if result.Err != nil {
		if strings.Contains(result.Err.Error(), "401") {
			h.Logger.Infof("Login options returned 401. This may happen when your c8y hostname is not setup correctly")
			return nil
		}
		h.Logger.Warnf("Failed to get login options. %s", result.Err)
		return result.Err
	}

	options := make([]loginOption, 0)
	for option := range result.Seq() {
		options = append(options, loginOption{
			Type:        option.Type(),
			InitRequest: option.InitRequest(),
			TFAStrategy: option.TFAStrategy(),
			Self:        option.Self(),
		})
	}

	if h.auth.Tenant != "" && len(options) > 0 {
		tenantSelfURL := options[0].Self
		if strings.HasPrefix(tenantSelfURL, "http") && !strings.Contains(tenantSelfURL, h.auth.Tenant+".") {
			h.Logger.Warningf("Detected invalid tenant name. expected %s to include %s. Tenant name will be ignored", h.auth.Tenant, tenantSelfURL)
			h.auth.Tenant = ""
			h.syncAuth()
		}
	}

	h.loginOptions = options
	h.sortLoginOptions()

	if len(h.loginOptions) > 0 {
		h.LoginType = h.loginOptions[0].Type

		// Setting preferred login method
		h.Logger.Debugf("Preferred login method. type=%s", h.LoginType)
	}
	return nil
}

func (h *LoginHandlerV2) sortLoginOptions() {
	if h.loginOptions == nil {
		return
	}

	optionOrder := map[string]int{
		logintype.None:           40,
		logintype.Basic:          30,
		logintype.OAuth2Internal: 20,

		// reduce priority due to common misconfiguration and
		// currently only the OAUTH2 Device Flow works and not many users have this setup correctly
		logintype.OAuth2: 50,
	}

	if h.LoginType != "" {
		if _, ok := optionOrder[h.LoginType]; ok {
			h.Logger.Infof("Setting preferred login method. %s", h.LoginType)
			optionOrder[h.LoginType] = -999 // Try preferred method first
		} else {
			h.Logger.Infof("Unsupported login method. The given option will be ignored. %s", h.LoginType)
		}
	}

	// Optionally limit the login types based on user settings
	if len(h.AllowedLoginTypes) > 0 {
		// only include the user-given login type
		h.Logger.Debugf("Filtering the login options based on a user-defined list. types=%v", h.AllowedLoginTypes)
		allowedLoginOptions := make([]loginOption, 0, 1)
		allLoginTypes := make([]string, 0, len(h.loginOptions))
		matchingTypes := make([]string, 0)
		for _, option := range h.loginOptions {
			for _, allowed := range h.AllowedLoginTypes {
				allLoginTypes = append(allLoginTypes, option.Type)
				if strings.EqualFold(option.Type, strings.TrimSpace(allowed)) {
					allowedLoginOptions = append(allowedLoginOptions, option)
					matchingTypes = append(matchingTypes, option.Type)
				}
			}
		}

		h.loginOptions = allowedLoginOptions
		h.Logger.Infof("Filtered login types. matches=%v. availableOnTenant=%v", matchingTypes, allLoginTypes)
	}

	// sort login options
	sort.SliceStable(h.loginOptions, func(i, j int) bool {
		iWeight := 100
		jWeight := 200

		if v, ok := optionOrder[h.loginOptions[i].Type]; ok {
			iWeight = v
		}

		if v, ok := optionOrder[h.loginOptions[j].Type]; ok {
			jWeight = v
		}
		return iWeight < jWeight
	})

	// log the preference of login types
	preferenceLoginTypes := make([]string, 0, len(h.loginOptions))
	for _, option := range h.loginOptions {
		preferenceLoginTypes = append(preferenceLoginTypes, option.Type)
	}

	h.Logger.Infof("Login type preference: %v", preferenceLoginTypes)
}

func (h *LoginHandlerV2) promptForPassword() error {
	validate := func(input string) error {
		if input == "" {
			return fmt.Errorf("Empty password")
		}
		return nil
	}

	if h.auth.Username == "" {
		prompter := prompt.NewPrompt(h.Logger)
		username, err := prompter.Username("Enter username", " ")
		if err != nil {
			return fmt.Errorf("user cancelled prompt")
		}
		h.auth.Username = username
	}

	reason := ""
	label := fmt.Sprintf("Enter c8y password for user (%s)", h.auth.Username)

	// Provide additional information to user what happened
	// invalid encryption key? or missing password
	if h.auth.Password == "" {
		reason = "password is empty"
	} else {
		reason = "password is invalid"
		label = fmt.Sprintf("Re-enter c8y password for user (%s)", h.auth.Username)
	}

	passwordPrompt := promptui.Prompt{
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
		passwordPrompt.Templates.Invalid = fmt.Sprintf("{{ \"✗ (%s)\" | red | faint }} {{ . | bold }}: {{ \"[input is hidden]\" | faint }}", reason)
		passwordPrompt.Templates.Valid = fmt.Sprintf("{{ \"✗ (%s)\" | red | faint }} {{ . | bold }}: {{ \"[input is hidden]\" | faint }}", reason)
	}
	pass, err := passwordPrompt.Run()

	if err != nil {
		return err
	}
	h.auth.Password = pass
	h.syncAuth()
	return nil
}

func (h *LoginHandlerV2) login() (LoginState, error) {
	if len(h.loginOptions) == 0 {
		// Don't fail on no login options. Default to using BASIC auth
		h.Logger.Infof("No login options available. Using BASIC auth")
		if h.authorized {
			return LoginStateAuth, nil
		}
		h.LoginType = logintype.Basic
		return LoginStateVerify, nil
	}

	if h.authorized {
		return LoginStateAuth, nil
	}

	h.loginAttempted = true

	order := []string{}
	for _, option := range h.loginOptions {
		order = append(order, option.Type)
	}
	h.Logger.Infof("Login type order: %v", order)

	for _, option := range h.loginOptions {
		h.Logger.Infof("Trying login option. %s", option.Type)
		switch option.Type {
		case logintype.OAuth2:
			// Device Authorization Flow (for external providers)
			if !h.IO.CanPromptOnStdErr() {
				err := fmt.Errorf("OAuth2 device flow requires an interactive console")
				return LoginStateAbort, err
			}

			// Note: The flow replaces the authorization of the client with the token
			deviceFlow := h.deviceFlowOptions()
			accessToken, loginErr := h.V2Client.AuthorizeWithDeviceFlow(h.ctx(), option.InitRequest, deviceFlow.AuthEndpoints, deviceFlow.DisplayFunc)
			if loginErr != nil {
				// Ignore SSO if invalid configuration is found
				if errors.Is(loginErr, apiv2.ErrSSOInvalidConfiguration) {
					h.Logger.Warnf("Skipping login type (%s) as SSO configuration is invalid. err=%s", option.Type, loginErr)
					continue
				}

				h.Logger.Warnf("Skipping login type (%s) as SSO login failed. err=%s", option.Type, loginErr)
				continue
			}

			h.Logger.Infof("Received access token via device flow. type=%s, scope=%s, tokenPresent=%v, refreshTokenPresent=%v", accessToken.Type, accessToken.Scope, accessToken.Token != "", accessToken.RefreshToken != "")
			h.auth.Token = accessToken.Token
			h.authType = authv2.AuthTypeBearer
			h.syncAuth()
			h.onSave()
			return LoginStateVerify, nil

		case logintype.OAuth2Internal:
			if h.authorized {
				h.onSave()
				return LoginStateVerify, nil
			}

			h.Logger.Infof("Logging in using %s", logintype.OAuth2Internal)

			// Check if username/password are provided
			if h.auth.Username == "" {
				h.Logger.Warnf("Skipping login type (%s) as a username is empty", option.Type)
				continue
			}
			if h.auth.Password == "" {
				h.Logger.Warnf("Skipping login type (%s) as password is empty", option.Type)
				continue
			}

			if err := h.loginInternal(); err != nil {
				if h.abortErr != nil {
					// e.g. the TFA setup needs a terminal, or the user cancelled the prompt
					return LoginStateAbort, h.abortErr
				}

				// Check for known message that the server responds with
				// when TFA is required, so don't log it on the ERROR level as it is annoying for users
				message := errorMessage(err)
				if strings.Contains(message, `For input string: "undefined"`) {
					h.Logger.Infof("OAuth2 most likely requires TFA. %s", message)
				} else {
					h.Logger.Errorf("OAuth2 failed. %s", message)
				}

				// trigger unknown to recheck if the credentials are valid
				return LoginStateUnknown, nil
			}
			h.authorized = true
			h.onSave()
			return LoginStateAuth, nil

		case logintype.Basic:
			if h.auth.Username == "" || h.auth.Password == "" {
				if h.IO.CanPromptOnStdErr() {
					return LoginStatePromptPassword, nil
				}
				return LoginStateAbort, fmt.Errorf("username or password is empty and the interactive prompt is disabled")
			}
			return LoginStateVerify, nil
		}
	}

	// return an error by default
	return LoginStateAbort, fmt.Errorf("no valid login type found")
}

// loginInternal logs in using the credentials of the user and requests a token (OAUTH2_INTERNAL).
// The two-factor authentication (setup, TOTP and SMS) and forced password changes are handled
// by the SDK, which calls the prompts of the handler when they are needed.
func (h *LoginHandlerV2) loginInternal() error {
	// The login requires the credentials, and the token must not be used
	h.applyAuth(authv2.AuthOptions{
		Tenant:   h.auth.Tenant,
		Username: h.auth.Username,
		Password: h.auth.Password,
	})

	opts := apiv2.LoginOptions{
		Method:           authv2.LoginMethodOAuth2Internal,
		TOTPSecret:       h.TOTPSecret,
		TOTPCode:         h.promptTOTPCode,
		QRCode:           h.displayTOTPQRCode,
		SMSCode:          h.promptSMSCode,
		PasswordChange:   h.promptPasswordChange,
		CredentialPrompt: h.promptCredentials,
	}
	token, err := h.V2Client.LoginWithOptions(h.ctx(), opts)
	if err != nil {
		h.syncAuth()
		return err
	}

	// the login can change the credentials, e.g. the password after a forced password change
	h.auth.Username = h.V2Client.Auth.Username
	h.auth.Password = h.V2Client.Auth.Password
	h.auth.Token = token
	h.authType = authv2.AuthTypeBearer
	h.syncAuth()

	// check that the token works, and get the tenant and the Cumulocity version
	tenant := h.V2Client.Tenants.Current.Get(h.ctx(), currenttenant.GetOptions{})
	if tenant.Err != nil {
		return tenant.Err
	}
	h.version = h.fetchVersion()
	h.auth.Tenant = tenant.Data.Name()
	h.syncAuth()
	return nil
}

// fetchVersion gets the Cumulocity version, but doesn't fail if it does not work
func (h *LoginHandlerV2) fetchVersion() string {
	result := h.V2Client.Tenants.SystemOptions.Get(h.ctx(), systemoptions.GetOption{
		Category: "system",
		Key:      "version",
	})
	if result.Err != nil {
		h.Logger.Warnf("Could not get Cumulocity System version. %s", result.Err)
		return h.version
	}
	return result.Data.Value()
}

func (h *LoginHandlerV2) verify() (LoginState, error) {
	h.attempts++

	if h.attempts > 2 {
		return LoginStateAbort, fmt.Errorf("max log attempts reached")
	}

	tenant := h.V2Client.Tenants.Current.Get(h.ctx(), currenttenant.GetOptions{})

	if tenant.Err != nil && apiv2.ErrHasStatus(tenant.Err, http.StatusUnauthorized) {
		message := errorMessage(tenant.Err)
		h.Logger.Infof("error message from server. %s", message)

		switch {
		case errorContains(message, "TFA TOTP setup required"):
			h.tfaCodeRequired = true
			return LoginStateTFASetup, nil
		case errorContains(message, "TFA TOTP code required"):
			h.Logger.Infof("TFA code is required. server response: %s", message)
			h.tfaCodeRequired = true
			return LoginStateNoAuth, nil
		case errorContains(message, "User has been logged out"):
			h.Logger.Warning("User had been logged out. Clearing token and trying again")
			h.auth.Token = ""
			h.syncAuth()
			h.onSave()
			return LoginStateNoAuth, nil
		case errorContains(message, "Tenant has no access from outside the platform"):
			// Decide what to do here
			h.LoginType = logintype.Basic
			return LoginStateVerify, nil
		case errorContains(message, "Bad credentials") || errorContains(message, "Invalid credentials"):
			h.Logger.Infof("Bad credentials, using auth method: %s", h.authType)

			// try resetting the tenant (in case if it is incorrect)
			h.auth.Tenant = ""
			h.syncAuth()

			if h.authType != authv2.AuthTypeBearer {
				if h.IO.CanPromptOnStdErr() {
					return LoginStatePromptPassword, nil
				}
				return LoginStateAbort, nil
			}
			return LoginStateNoAuth, nil
		default:
			return LoginStateNoAuth, nil
		}
	}

	if tenant.Err != nil {
		return LoginStateAuth, tenant.Err
	}

	h.Logger.Infof("Detected tenant: %s", tenant.Data.Name())
	h.auth.Tenant = tenant.Data.Name()
	h.syncAuth()

	// Get Cumulocity system version
	if h.version == "" {
		h.version = h.fetchVersion()
	}

	// Get user information if not set (e.g. external SSO tokens don't generally include the username in a known field)
	if h.auth.Username == "" {
		h.Logger.Infof("Getting the current user's information as the username is not set")
		currentUser := h.V2Client.Users.CurrentUser.Get(h.ctx())
		if currentUser.Err != nil {
			h.Logger.Warnf("Could not get username. %s", currentUser.Err)
		} else {
			h.Logger.Infof("Found current username. %s", currentUser.Data.UserName())
			h.auth.Username = currentUser.Data.UserName()
			h.syncAuth()
		}
	}
	return LoginStateAuth, nil
}
