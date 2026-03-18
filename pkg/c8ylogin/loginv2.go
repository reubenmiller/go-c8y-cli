// Package c8ylogin provides authentication flows for Cumulocity IoT.
//
// This file (loginv2.go) integrates the go-c8y v2 SDK's authentication
// mechanisms into the CLI while keeping full compatibility with the existing
// *c8y.Client type used throughout the rest of the codebase.
//
// Both modules coexist in go.mod:
//
//	github.com/reubenmiller/go-c8y       – existing client (pkg/c8y)
//	github.com/reubenmiller/go-c8y/v2    – new auth API   (pkg/c8y/api)
//
// The *c8y.Client field on LoginHandlerV2 uses the old module so that call
// sites need only change the constructor; all field accesses remain identical.
// Once the full migration to the v2 module is complete, the old dependency
// can be dropped and the two imports collapsed into one.
//
// To add the v2 module:
//
//	go get github.com/reubenmiller/go-c8y/v2
//
// # New capabilities unlocked by the v2 SDK
//
//   - CERTIFICATE: mTLS device-certificate authentication (no username / password).
//     The client exchanges a PEM certificate + private key for a short-lived
//     bearer token via the Cumulocity /devicecontrol/deviceAccessToken endpoint.
//
//   - OAUTH2_BROWSER_FLOW: Authorization Code flow (RFC 6749) via the system
//     browser. A local callback server is started automatically; the browser is
//     opened to the IdP login page; the returned code is exchanged for a token.
//     Supports PKCE (RFC 7636) for future compatibility.
//
//   - OAUTH2_DEVICE_FLOW: unchanged semantics compared to v1, but now uses
//     auto-discovery of OAuth2 endpoints via OpenID Connect configuration.
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
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cli/browser"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/mdp/qrterminal/v3"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logintype"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/prompt"

	// pkg/c8y – the existing c8y.Client type used throughout the CLI.
	// Uses the old module so *c8y.Client remains type-compatible with call sites.
	"github.com/reubenmiller/go-c8y/pkg/c8y"

	// pkg/c8y/api – the new authentication API introduced in the v2 SDK.
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	authv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/currenttenant"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/users/currentuser/totp"
	oauth2api "github.com/reubenmiller/go-c8y/v2/pkg/oauth/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/oauth/device"
)

// LoginHandlerV2 is a drop-in replacement for LoginHandler that drives all
// login flows through the go-c8y v2 SDK. It supports every authentication
// method of the original handler and additionally exposes:
//
//   - Certificate (mTLS device-certificate)
//   - OAUTH2 Browser Flow (Authorization Code via system browser)
//   - Automatic TOTP code generation from a pre-shared secret
//   - Structured callbacks for every interactive prompt
//
// After a successful login the obtained bearer token is synced back into the
// existing *c8y.Client so that the rest of the CLI continues working without
// modification.

// LoginTypeBrowser, LoginTypeDevice and LoginTypeCertificate are re-exported
// from pkg/logintype for the convenience of callers that already import c8ylogin
// and want to avoid a second import.
const (
	LoginTypeBrowser     = logintype.Browser
	LoginTypeDevice      = logintype.Device
	LoginTypeCertificate = logintype.Certificate
)

// # Minimal call-site migration from LoginHandler
//
//	// Before (v1)
//	handler := c8ylogin.NewLoginHandler(io, client, w, onSave)
//	handler.LoginType = loginType           // e.g. "OAUTH2_INTERNAL"
//	handler.SSO = ssoSettings
//	err = handler.Run()
//	session.Username  = handler.C8Yclient.Username
//	session.LoginType = handler.LoginType
//
//	// After (v2)
//	handler := c8ylogin.NewLoginHandlerV2(io, client, w, onSave)
//	handler.LoginType = loginType           // same string, automatically converted
//	handler.SSO = ssoSettings
//	err = handler.Run()
//	session.Username  = handler.C8Yclient.Username  // unchanged – same v1 field name
//	session.LoginType = handler.LoginType           // updated after login
type LoginHandlerV2 struct {
	IO     *iostreams.IOStreams
	Writer io.Writer
	Logger *logger.Logger

	// V2Client is the go-c8y v2 api.Client used for all login operations.
	// Created automatically by NewLoginHandlerV2; can be replaced for testing.
	V2Client *apiv2.Client

	// C8Yclient is the existing c8y.Client (pkg/c8y) used throughout the CLI,
	// kept under the same field name as LoginHandler so call sites can read
	// handler.C8Yclient.Username / .BaseURL / .Token after Run() without any
	// changes. After a successful login its Token field is updated with the
	// newly obtained token so subsequent API calls keep working.
	C8Yclient *c8y.Client

	// LoginType is the login type string (e.g. "OAUTH2_INTERNAL",
	// "BASIC", "OAUTH2"). It is:
	//   - accepted as input and auto-converted to LoginMethod / LoginTypePreference
	//   - updated after a successful Run() to reflect the method that was used
	//
	// v2-only methods:
	//   "DEVICE"       → LoginMethodOAuth2DeviceFlow  (explicit alias; "OAUTH2" also accepted)
	//   "BROWSER"      → LoginMethodOAuth2BrowserFlow
	//   "CERTIFICATE"  → LoginMethodCertificate
	LoginType string

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

	onSave func()
}

// NewLoginHandlerV2 creates a LoginHandlerV2 from an existing *c8y.Client.
// It constructs an api.Client (pkg/c8y/api) using the same host, tenant,
// username, password, and token so both clients share the same session context.
func NewLoginHandlerV2(
	IO *iostreams.IOStreams,
	v1 *c8y.Client,
	w io.Writer,
	onSave func(),
) *LoginHandlerV2 {
	h := &LoginHandlerV2{
		IO:        IO,
		C8Yclient: v1,
		Writer:    w,
		onSave:    onSave,
		Logger:    logger.NewDummyLogger("c8yloginv2"),
	}

	h.V2Client = apiv2.NewClient(apiv2.ClientOptions{
		BaseURL: v1.BaseURL.String(),
		Auth: authv2.AuthOptions{
			Tenant:   v1.TenantName,
			Username: v1.Username,
			Password: v1.Password,
			Token:    v1.Token,
		},
	})

	return h
}

// SetLogger replaces the no-op logger with a real one.
func (h *LoginHandlerV2) SetLogger(l *logger.Logger) { h.Logger = l }

// Run performs the full login flow and syncs the resulting token back into
// the v1 client. It is a drop-in replacement for LoginHandler.Run.
//
// LoginType (v1-style string) is resolved to Method / Preference just before
// the login attempt so callers can set either field.
func (h *LoginHandlerV2) Run() error {
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
			// Update the v1-style LoginType string so call sites that read
			// handler.LoginType after Run() see the correct value.
			h.LoginType = loginTypeFromMethod(m)
		},
	}

	token, err := h.V2Client.LoginWithOptions(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Sync token into C8Yclient so the rest of the CLI works as
	// before without any change to code outside this package.
	if token != "" && h.C8Yclient != nil {
		h.C8Yclient.SetToken(token)
	}

	// Populate the tenantID so it is exposed in the env
	tenantResult := h.V2Client.Tenants.Current.Get(context.Background(), currenttenant.GetOptions{})
	if tenantResult.Err == nil {
		if tenantName := tenantResult.Data.Name(); tenantName != "" {
			h.C8Yclient.TenantName = tenantName
		}
	}

	h.onSave()
	return nil
}

// loginTypeFromMethod converts a v2 LoginMethod back to the v1-style string so
// that existing call-site code that reads handler.LoginType continues working.
func loginTypeFromMethod(m authv2.LoginMethod) string {
	switch m {
	case authv2.LoginMethodBasic:
		return c8y.LoginTypeBasic
	case authv2.LoginMethodOAuth2Internal:
		return c8y.LoginTypeOAuth2Internal
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

// preferenceFromLoginType converts a v1-style login type string (e.g.
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
//	"CERTIFICATE"      → [CERTIFICATE]           (new in v2)
//	"BROWSER"          → [OAUTH2_BROWSER_FLOW]   (new in v2)
func PreferenceFromLoginType(loginType string) []authv2.LoginMethod {
	switch loginType {
	case c8y.LoginTypeOAuth2Internal:
		return []authv2.LoginMethod{authv2.LoginMethodOAuth2Internal}
	case c8y.LoginTypeBasic, c8y.LoginTypeNone:
		return []authv2.LoginMethod{authv2.LoginMethodBasic}
	case logintype.Device, c8y.LoginTypeOAuth2:
		// logintype.Device is the preferred explicit name; "OAUTH2" kept as a legacy alias.
		return []authv2.LoginMethod{authv2.LoginMethodOAuth2DeviceFlow}
	case logintype.Certificate:
		// New auth type – not present in v1 constants.
		return []authv2.LoginMethod{authv2.LoginMethodCertificate}
	case logintype.Browser:
		// New auth type – not present in v1 constants.
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
			if len(input) < 6 {
				return fmt.Errorf("code must be at least 6 digits")
			}
			return nil
		},
	}
	return p.Run()
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
				return fmt.Errorf("user cancelled username prompt")
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
