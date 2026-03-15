// Package logintype defines the canonical string constants for every login
// type recognised by the CLI. These values are used in --loginType flags,
// session files, the settings store, and tab-completion lists.
//
// This package has no dependencies beyond the standard library so that it can
// be safely imported by any package in the tree without creating import cycles.
package logintype

const (
	// Basic selects HTTP Basic authentication on every request.
	Basic = "BASIC"

	// OAuth2Internal exchanges username/password for an OAI-Secure token.
	OAuth2Internal = "OAUTH2_INTERNAL"

	// OAuth2 selects an external OAUTH2/SSO provider (device or legacy flow).
	OAuth2 = "OAUTH2"

	// None disables authentication entirely.
	None = "NONE"

	// Browser selects the Authorization Code flow via the system browser.
	// Maps to go-c8y/v2 LoginMethodOAuth2BrowserFlow ("OAUTH2_BROWSER_FLOW").
	Browser = "BROWSER"

	// Device selects the RFC 8628 device authorization flow.
	// Maps to go-c8y/v2 LoginMethodOAuth2DeviceFlow ("OAUTH2_DEVICE_FLOW").
	Device = "DEVICE"

	// Certificate selects mTLS device-certificate authentication.
	// Maps to go-c8y/v2 LoginMethodCertificate ("CERTIFICATE").
	Certificate = "CERTIFICATE"
)
