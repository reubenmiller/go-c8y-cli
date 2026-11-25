// Package c8ylogin provides the authentication flows for Cumulocity.
//
// All flows use the go-c8y v2 SDK (pkg/c8y/api). The handler supports:
//
//   - BASIC and OAUTH2_INTERNAL (OAI-Secure tokens, including TOTP/SMS two-factor
//     authentication and forced password changes), where the tenant's login options
//     decide which login type is used (see legacyflow.go)
//   - OAUTH2 (device flow with the legacy login type)
//   - DEVICE: OAuth2 device flow
//   - BROWSER: OAuth2 authorization code flow via the system browser
//   - CERTIFICATE: mTLS device-certificate authentication (no username / password)
//
// Further capabilities of the SDK which are used by the flows:
//
//   - Automatic TOTP (TOTPSecret field): CI / automation scenarios can supply a
//     base-32 TOTP secret so that two-factor codes are generated automatically
//     instead of prompting the user.
//
//   - Callback-based credential prompting (CredentialPrompt): instead of
//     embedding prompt logic inside the login state machine the caller supplies
//     a function that is invoked when credentials are missing, keeping the UI
//     layer fully decoupled from the auth logic.
//
//   - Forced-password-change handling (PasswordChange callback): the server can
//     require the user to set a new password on first login; the v2 SDK surfaces
//     this as a structured callback rather than an ad-hoc error path.

package c8ylogin

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/cli/browser"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/mdp/qrterminal/v3"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logintype"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/prompt"
	"golang.org/x/term"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	authv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/currenttenant"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/users/currentuser/totp"
	oauth2api "github.com/reubenmiller/go-c8y/v2/pkg/oauth/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/oauth/device"
)

var (
	// Timeout in milliseconds used for all Server requests
	Timeout int = 30000
)

// LoginTypeBrowser, LoginTypeDevice and LoginTypeCertificate are re-exported
// from pkg/logintype for the convenience of callers that already import c8ylogin
// and want to avoid a second import.
const (
	LoginTypeBrowser     = logintype.Browser
	LoginTypeDevice      = logintype.Device
	LoginTypeCertificate = logintype.Certificate
)

// LoginHandlerV2 drives all login flows through the go-c8y v2 SDK.
//
// Usage:
//
//	handler := c8ylogin.NewLoginHandlerV2(io, apiv2.ClientOptions{BaseURL: host, Auth: auth}, w, onSave)
//	handler.LoginType = loginType           // e.g. "OAUTH2_INTERNAL"
//	handler.SSO = ssoSettings
//	err = handler.Run()
//	session := handler.Session()            // the credentials after the login
//	session.LoginType = handler.LoginType   // updated after login
type LoginHandlerV2 struct {
	IO     *iostreams.IOStreams
	Writer io.Writer
	Logger *logger.Logger

	// V2Client is the go-c8y v2 api.Client used for all login operations.
	// It is created by NewLoginHandlerV2 and is replaced when the authorization of
	// the handler changes, so don't keep a reference to it. Use Session() to get the
	// credentials after the login.
	V2Client *apiv2.Client

	// LoginType is the login type string (e.g. "OAUTH2_INTERNAL",
	// "BASIC", "OAUTH2"). It is:
	//   - accepted as input and auto-converted to LoginMethod / LoginTypePreference
	//   - updated after a successful Run() to reflect the method that was used
	//
	// Additional methods:
	//   "DEVICE"       → LoginMethodOAuth2DeviceFlow  (explicit alias; "OAUTH2" also accepted)
	//   "BROWSER"      → LoginMethodOAuth2BrowserFlow
	//   "CERTIFICATE"  → LoginMethodCertificate
	LoginType string

	// AllowedLoginTypes is a user defined list of the allowed login types (from their perspective).
	// It limits the login options of the tenant which are tried. This is only used for the
	// BASIC / OAUTH2_INTERNAL / OAUTH2 login types.
	AllowedLoginTypes []string

	// LoginMethod enforces a specific login flow. When set it takes precedence
	// over LoginTypePreference. Leave empty to use the tenant's advertised
	// login options, prioritised by the default Preference list.
	//
	// Usually you set LoginType instead; LoginMethod is for callers that already
	// work with v2 LoginMethod constants directly.
	LoginMethod authv2.LoginMethod

	// LoginTypePreference is an ordered list of login methods to attempt.
	// Derived automatically from LoginType via PreferenceFromLoginType;
	// can be overridden directly when finer control is needed.
	LoginTypePreference []authv2.LoginMethod

	// LastUsedMethod is set by Run() to the LoginMethod that succeeded.
	// Useful for callers that want to persist the winning method.
	LastUsedMethod authv2.LoginMethod

	// SSO settings forwarded to the Device / Browser flows.
	SSO config.SSOSettings

	// BrowserFlow controls the Authorization Code flow (opened in the system
	// browser). When nil, sensible defaults are used (127.0.0.1:5001/callback).
	// Set this to override any option; BrowserCallbackURL is a shorthand for
	// setting only the CallbackURL without constructing the full struct.
	BrowserFlow *apiv2.BrowserFlowOptions

	// BrowserCallbackURL overrides the default redirect URI
	// (http://127.0.0.1:5001/callback) used by the browser flow local server.
	// The value must exactly match a URI pre-registered in the SSO provider.
	// Accepted forms:
	//   http://127.0.0.1:5001/callback   – full URI
	//   127.0.0.1:5001/callback          – scheme inferred as http
	//   127.0.0.1:5001                   – path defaults to /callback
	// Ignored when BrowserFlow is set explicitly.
	BrowserCallbackURL string

	// CertificatePath is the path to the PEM-encoded client certificate file.
	// Required when LoginMethod == LoginMethodCertificate.
	// Can also be set via config; the CredentialPrompt callback will ask
	// the user interactively if both fields are empty.
	CertificatePath string

	// CertificateKeyPath is the path to the PEM-encoded private key file.
	// Required when LoginMethod == LoginMethodCertificate.
	CertificateKeyPath string

	// TOTPSecret is an optional base-32 TOTP secret for automation scenarios.
	// When supplied, TOTP challenge codes are generated automatically (RFC 6238)
	// instead of prompting the user interactively. Do NOT set this for
	// interactive sessions – it negates the second-factor security benefit.
	TOTPSecret string

	// TOTPCode optional two-factor code entered by the user elsewhere (e.g. in the
	// browser by the MCP server, or by the --tfaCode flag). It is used once instead
	// of a prompt
	TOTPCode     string
	totpCodeUsed bool

	onSave func()

	// clientOpts are the options used to (re)create the client
	clientOpts apiv2.ClientOptions

	// auth the credentials of the session
	auth authv2.AuthOptions

	// authType the authorization which is used for the requests
	authType authv2.AuthType

	// version the Cumulocity version
	version string

	// state of the login flow for the BASIC / OAUTH2_INTERNAL / OAUTH2 login types
	loginOptions    []loginOption
	attempts        int
	authorized      bool
	tfaCodeRequired bool
	loginAttempted  bool

	// abortErr is the error of a prompt that must stop the login (e.g. no terminal)
	abortErr error
}

// NewLoginHandlerV2 creates a LoginHandlerV2 which uses a go-c8y v2 client created from the given options.
// The credentials of the options (opts.Auth) are used as the session credentials, and opts.Auth.AuthType
// selects the authorization which is used for the requests.
//
// The client is created without the credentials, so no token is requested before the login
// starts. Set opts.Transport / opts.InsecureSkipVerify to control the connection.
func NewLoginHandlerV2(
	IO *iostreams.IOStreams,
	opts apiv2.ClientOptions,
	w io.Writer,
	onSave func(),
) *LoginHandlerV2 {
	h := &LoginHandlerV2{
		IO:     IO,
		Writer: w,
		onSave: onSave,
		Logger: logger.NewDummyLogger("c8yloginv2"),
	}
	if h.onSave == nil {
		h.onSave = func() {}
	}

	h.auth = opts.Auth
	h.authType = opts.Auth.GetAuthTypes()[0]

	// The credentials are applied after the client is created. Otherwise the client requests a
	// token as soon as it has a username and password, before the login flow had a chance to
	// handle two-factor authentication.
	opts.Auth = authv2.AuthOptions{Tenant: opts.Auth.Tenant}
	opts.SkipTenantLookup = true
	if opts.Timeout == 0 {
		opts.Timeout = time.Duration(Timeout) * time.Millisecond
	}
	if opts.Logger == nil {
		var errOut io.Writer = io.Discard
		if IO != nil && IO.ErrOut != nil {
			errOut = IO.ErrOut
		}
		opts.Logger = slog.New(slog.NewTextHandler(errOut, &slog.HandlerOptions{Level: slog.LevelWarn}))
	}
	h.clientOpts = opts
	h.applyAuth(h.requestAuth())
	return h
}

// SetLogger replaces the no-op logger with a real one.
func (h *LoginHandlerV2) SetLogger(l *logger.Logger) { h.Logger = l }

// Session returns the credentials of the session. The values are updated by the login,
// e.g. with the token and the tenant.
func (h *LoginHandlerV2) Session() config.ClientSession {
	session := config.NewClientSession(h.V2Client.BaseURL, h.auth, h.version)
	session.AuthType = h.authType
	return session
}

// applyAuth replaces the client with a new client which uses the given authorization
func (h *LoginHandlerV2) applyAuth(auth authv2.AuthOptions) {
	client := apiv2.NewClient(h.clientOpts)
	client.SetAuth(auth)
	h.V2Client = client
}

// requestAuth returns the authorization which is used for requests, based on the authorization type
func (h *LoginHandlerV2) requestAuth() authv2.AuthOptions {
	auth := authv2.AuthOptions{
		Tenant: h.auth.Tenant,
	}
	authType := h.authType
	if authType == authv2.AuthTypeUnset {
		if h.auth.Token != "" {
			authType = authv2.AuthTypeBearer
		} else if h.auth.Username != "" && h.auth.Password != "" {
			authType = authv2.AuthTypeBasic
		}
	}
	switch authType {
	case authv2.AuthTypeBearer:
		auth.Token = h.auth.Token
	case authv2.AuthTypeBasic:
		auth.Username = h.auth.Username
		auth.Password = h.auth.Password
	}
	return auth
}

// syncAuth updates the client after the credentials or the authorization type has changed
func (h *LoginHandlerV2) syncAuth() {
	h.applyAuth(h.requestAuth())
}

// ctx returns the context used for requests
func (h *LoginHandlerV2) ctx() context.Context {
	return context.Background()
}

// isSDKFlow checks if the login is done using the login flows of the SDK
// (instead of the login options of the tenant)
func (h *LoginHandlerV2) isSDKFlow() bool {
	if h.LoginMethod != "" || len(h.LoginTypePreference) > 0 {
		return true
	}
	switch strings.ToUpper(h.LoginType) {
	case logintype.Browser, logintype.Certificate, logintype.Device:
		return true
	}
	return false
}

// Run performs the full login flow and stores the resulting credentials in the handler.
// Use Session() to get the credentials after the login.
//
// LoginType (string) is resolved to Method / Preference just before
// the login attempt so callers can set either field.
func (h *LoginHandlerV2) Run() error {
	if h.isSDKFlow() {
		return h.runSDKFlow()
	}
	return h.runTenantFlow()
}

func (h *LoginHandlerV2) runSDKFlow() error {
	// The login needs the credentials, and not only the ones used by the authorization type
	h.applyAuth(h.auth)

	// Resolve LoginType string → v2 method / preference.
	// LoginMethod takes precedence; otherwise derive from LoginType string.
	method := h.LoginMethod
	preference := h.LoginTypePreference
	if method == "" && len(preference) == 0 && h.LoginType != "" {
		preference = PreferenceFromLoginType(h.LoginType)
	}

	opts := apiv2.LoginOptions{
		// -- Method selection ------------------------------------------------
		// Method and Preference are mutually exclusive. When Method is set the
		// v2 SDK enforces that single flow; when only Preference is set the
		// first locally-available method in the list wins.
		Method:     method,
		Preference: preference,

		// -- TOTP / TFA ------------------------------------------------------
		// TOTPSecret enables automatic code generation for CI pipelines.
		// When empty the user is prompted interactively via TOTPCode.
		TOTPSecret: h.TOTPSecret,
		TOTPCode:   h.promptTOTPCode,
		QRCode:     h.displayTOTPQRCode,

		// -- Other challenges ------------------------------------------------
		SMSCode:          h.promptSMSCode,
		PasswordChange:   h.promptPasswordChange,
		CredentialPrompt: h.promptCredentials,

		// -- Flow-specific options -------------------------------------------
		BrowserFlow: h.browserFlowOptions(),
		DeviceFlow:  h.deviceFlowOptions(),

		// -- Post-login hook -------------------------------------------------
		OnSuccess: func(m authv2.LoginMethod) {
			h.Logger.Infof("Login succeeded. method=%s", m)
			h.LastUsedMethod = m
			// Update the LoginType string so call sites that read
			// handler.LoginType after Run() see the correct value.
			h.LoginType = loginTypeFromMethod(m)
		},
	}

	token, err := h.V2Client.LoginWithOptions(h.ctx(), opts)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Keep the credentials which the login has set (e.g. the token, or credentials from the prompts)
	h.auth = h.V2Client.Auth
	if token != "" {
		h.auth.Token = token
	}

	// Populate the tenantID so it is exposed in the env
	tenantResult := h.V2Client.Tenants.Current.Get(h.ctx(), currenttenant.GetOptions{})
	if tenantResult.Err == nil {
		if tenantName := tenantResult.Data.Name(); tenantName != "" {
			h.auth.Tenant = tenantName
		}
	}

	h.onSave()
	return nil
}

// loginTypeFromMethod converts a v2 LoginMethod back to the login type string so
// that existing call-site code that reads handler.LoginType continues working.
func loginTypeFromMethod(m authv2.LoginMethod) string {
	switch m {
	case authv2.LoginMethodBasic:
		return logintype.Basic
	case authv2.LoginMethodOAuth2Internal:
		return logintype.OAuth2Internal
	case authv2.LoginMethodOAuth2DeviceFlow:
		return logintype.Device
	case authv2.LoginMethodCertificate:
		return logintype.Certificate
	case authv2.LoginMethodOAuth2BrowserFlow:
		return logintype.Browser
	default:
		return string(m)
	}
}

// PreferenceFromLoginType converts a login type string (e.g.
// "OAUTH2_INTERNAL") into an ordered v2 LoginMethod preference list. This
// makes it easy to migrate existing --loginType flag values to v2.
//
// Mapping:
//
//	""                 → nil (v2 default: try OAUTH2_INTERNAL first)
//	"OAUTH2_INTERNAL"  → [OAUTH2_INTERNAL]
//	"BASIC"            → [BASIC]
//	"NONE"             → [BASIC]   (NONE means no SSO; fall back to BASIC)
//	"DEVICE"           → [OAUTH2_DEVICE_FLOW]   (preferred explicit name)
//	"OAUTH2"           → [OAUTH2_DEVICE_FLOW]   (legacy alias for DEVICE)
//	"CERTIFICATE"      → [CERTIFICATE]
//	"BROWSER"          → [OAUTH2_BROWSER_FLOW]
func PreferenceFromLoginType(loginType string) []authv2.LoginMethod {
	switch loginType {
	case logintype.OAuth2Internal:
		return []authv2.LoginMethod{authv2.LoginMethodOAuth2Internal}
	case logintype.Basic, logintype.None:
		return []authv2.LoginMethod{authv2.LoginMethodBasic}
	case logintype.Device, logintype.OAuth2:
		// logintype.Device is the preferred explicit name; "OAUTH2" kept as a legacy alias.
		return []authv2.LoginMethod{authv2.LoginMethodOAuth2DeviceFlow}
	case logintype.Certificate:
		return []authv2.LoginMethod{authv2.LoginMethodCertificate}
	case logintype.Browser:
		return []authv2.LoginMethod{authv2.LoginMethodOAuth2BrowserFlow}
	default:
		return nil
	}
}

// --------------------------------------------------------------------------
// Private helpers – flow option builders
// --------------------------------------------------------------------------

func (h *LoginHandlerV2) browserFlowOptions() *apiv2.BrowserFlowOptions {
	openBrowser := func(rawURL string) error {
		bold := color.New(color.Bold)
		bold.EnableColor()
		fmt.Fprintf(h.IO.ErrOut, "Opening the SSO login page in your browser so you can login...\n\n%s\n\n", bold.Sprint(rawURL))
		return browser.OpenURL(rawURL)
	}

	if h.BrowserFlow != nil {
		// Caller provided full options; just ensure OpenBrowser is set.
		if h.BrowserFlow.OpenBrowser == nil {
			h.BrowserFlow.OpenBrowser = openBrowser
		}
		return h.BrowserFlow
	}

	// Determine whether the browser flow is actually selected (either via
	// LoginMethod or via LoginType string).
	usesBrowserFlow := h.LoginMethod == authv2.LoginMethodOAuth2BrowserFlow ||
		strings.ToUpper(h.LoginType) == logintype.Browser
	if !usesBrowserFlow {
		return nil
	}

	return &apiv2.BrowserFlowOptions{
		OpenBrowser: openBrowser,
		CallbackURL: h.BrowserCallbackURL, // empty string → SDK uses its default
	}
}

func (h *LoginHandlerV2) deviceFlowOptions() *apiv2.DeviceFlowOptions {
	return &apiv2.DeviceFlowOptions{
		AuthEndpoints: oauth2api.AuthEndpoints{
			OpenIDConfigurationURL: h.SSO.DiscoveryURL,
			Scopes:                 h.SSO.Scopes,
			AuthRequestOptions: []oauth2api.AuthRequestEditorFn{
				oauth2api.WithAudience(h.SSO.Audience),
			},
		},
		DisplayFunc: func(code *device.CodeResponse) error {
			verificationURI := code.VerificationURIComplete
			if verificationURI == "" {
				verificationURI = code.VerificationURI
			}

			h.writeMessage("\nLogging in using the device flow (OAUTH2)\n\n")
			bold := color.New(color.Bold)
			bold.EnableColor()

			if code.VerificationURIComplete == "" {
				fmt.Fprintf(os.Stderr, " First copy your one-time code: %s\n", bold.Sprint(code.UserCode))
			}

			fmt.Fprintf(h.IO.ErrOut, "%s to open %s in your browser...", bold.Sprintf("Press Enter"), verificationURI)
			bufio.NewReader(h.IO.In).ReadBytes('\n') //nolint:errcheck

			if browserErr := browser.OpenURL(verificationURI); browserErr != nil {
				h.writeMessage("Failed to open browser, please open the URL in your browser manually\n")
			}
			return nil
		},
	}
}

// --------------------------------------------------------------------------
// Private helpers – interactive callbacks
// All of these are passed as function values to LoginOptions so they are only
// called when the SDK actually needs them.
// --------------------------------------------------------------------------

// promptTOTPCode is called by the SDK whenever a TOTP code is needed (either
// for first-time enrollment verification or for a regular login challenge).
func (h *LoginHandlerV2) promptTOTPCode(_ context.Context, challenge totp.TOTPChallenge) (string, error) {
	// a code which was given by the user (e.g. via a flag, or by the MCP server) is only used once,
	// so a rejected code isn't sent again. It can't be used to verify the setup, as the secret is new.
	if h.TOTPCode != "" && !h.totpCodeUsed && !challenge.IsSetup {
		h.totpCodeUsed = true
		return h.TOTPCode, nil
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		if challenge.IsSetup {
			h.abortErr = errors.New(config.TOTPSetupRequiredMessage)
		} else {
			h.abortErr = errors.New(config.TOTPCodeRequiredMessage)
		}
		return "", h.abortErr
	}

	if !challenge.IsSetup {
		h.Logger.Infof("TFA code is required")
		session := h.Session()
		os.Stderr.WriteString(fmt.Sprintf("Session details:\nHost=%s, username=%s\n", session.Host, session.Username))
	}

	label := "Enter Two-Factor code"
	if challenge.IsSetup {
		label = "Verify TOTP setup – enter the code shown in your authenticator app"
	}

	p := promptui.Prompt{
		Stdin:       os.Stdin,
		Stdout:      os.Stderr,
		HideEntered: true,
		Label:       label,
		Validate: func(input string) error {
			if len(strings.ReplaceAll(input, " ", "")) < 6 {
				return fmt.Errorf("code must be at least 6 digits")
			}
			return nil
		},
	}
	code, err := p.Run()
	if err != nil {
		h.abortErr = fmt.Errorf("User cancelled login")
		return "", h.abortErr
	}
	return code, nil
}

// displayTOTPQRCode is called during first-time TOTP enrollment so the user
// can scan the secret into their authenticator app.
func (h *LoginHandlerV2) displayTOTPQRCode(otpauthURL, rawSecret string) {
	qrterminal.GenerateWithConfig(otpauthURL, qrterminal.Config{
		Level:      qrterminal.M,
		Writer:     h.Writer,
		HalfBlocks: true,
		QuietZone:  1,
	})
	h.writeMessage("\n")
	fmt.Fprintf(h.Writer, "TOTP Secret: %s\n\n", rawSecret)
}

// promptSMSCode is called when the server has sent an SMS PIN to the user.
func (h *LoginHandlerV2) promptSMSCode(_ context.Context, challenge apiv2.SMSChallenge) (string, error) {
	fmt.Fprintf(h.Writer, "An SMS PIN has been sent to your registered phone. (%s)\n", challenge.Message)

	p := promptui.Prompt{
		Stdin:       os.Stdin,
		Stdout:      os.Stderr,
		HideEntered: true,
		Label:       "Enter SMS PIN",
		Validate: func(input string) error {
			if len(input) == 0 {
				return fmt.Errorf("PIN must not be empty")
			}
			return nil
		},
	}
	return p.Run()
}

// promptPasswordChange is called when the server forces the user to set a new
// password before allowing login.
func (h *LoginHandlerV2) promptPasswordChange(_ context.Context, challenge apiv2.PasswordChangeChallenge) (email, newPassword string, err error) {
	fmt.Fprintf(h.Writer, "Your account requires a password change before login.\n")

	prompter := prompt.NewPrompt(h.Logger)
	email, err = prompter.Input("Enter your e-mail address", challenge.Email, false, false)
	if err != nil {
		return
	}

	p := promptui.Prompt{
		Stdin:       os.Stdin,
		Stdout:      os.Stderr,
		HideEntered: true,
		Label:       "Enter new password",
		Validate: func(input string) error {
			if len(input) < 8 {
				return fmt.Errorf("password must be at least 8 characters")
			}
			return nil
		},
	}
	newPassword, err = p.Run()
	return
}

// promptCredentials is called when a credential-based method (BASIC,
// OAUTH2_INTERNAL, CERTIFICATE) cannot proceed because fields are empty. The
// function fills in whatever is missing so the SDK can retry.
func (h *LoginHandlerV2) promptCredentials(_ context.Context, method authv2.LoginMethod, auth *authv2.AuthOptions) error {
	prompter := prompt.NewPrompt(h.Logger)

	switch method {
	case authv2.LoginMethodBasic, authv2.LoginMethodOAuth2Internal:
		if auth.Username == "" {
			username, err := prompter.Username("Enter username", " ")
			if err != nil {
				h.abortErr = fmt.Errorf("user cancelled username prompt")
				return h.abortErr
			}
			auth.Username = username
		}
		if auth.Password == "" {
			p := promptui.Prompt{
				Stdin:       os.Stdin,
				Stdout:      os.Stderr,
				HideEntered: true,
				Pointer:     emptyPointer,
				Label:       fmt.Sprintf("Enter password for %s 🔒", auth.Username),
				Validate: func(input string) error {
					if input == "" {
						return fmt.Errorf("password must not be empty")
					}
					return nil
				},
			}
			pass, err := p.Run()
			if err != nil {
				return err
			}
			auth.Password = pass
		}

	case authv2.LoginMethodCertificate:
		// For certificate auth the paths are part of the handler config rather
		// than the auth options because they are file paths, not raw values.
		// Sync them into auth so the SDK can use them.
		if h.CertificatePath != "" {
			auth.Certificate = h.CertificatePath
		}
		if h.CertificateKeyPath != "" {
			auth.CertificateKey = h.CertificateKeyPath
		}
		if auth.Certificate == "" {
			certPath, err := prompter.Input("Path to client certificate (PEM)", "", true, false)
			if err != nil {
				return fmt.Errorf("user cancelled certificate prompt")
			}
			auth.Certificate = certPath
		}
		if auth.CertificateKey == "" {
			keyPath, err := prompter.Input("Path to certificate private key (PEM)", "", true, false)
			if err != nil {
				return fmt.Errorf("user cancelled key prompt")
			}
			auth.CertificateKey = keyPath
		}
	}

	return nil
}

func (h *LoginHandlerV2) writeMessage(m string) {
	if _, err := io.WriteString(h.Writer, m); err != nil {
		h.Logger.Warnf("Failed to write message. %s", err)
	}
}
