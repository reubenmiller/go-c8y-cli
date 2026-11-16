package extsdk

import (
	"io"
	"os"
)

// Session is a session which a session provider returns to c8y. It uses the same json format
// as c8y session files. Set Authorized to true if the credentials are already verified (for
// example a token which the provider requested itself), so c8y does not log in again.
//
// Declare the provider in extension.yaml:
//
//	sessionProviders:
//	  - name: myprovider
//	    description: Select a session from my vault
//	    exec: [sessions, select]
//	    secrets: [MY_VAULT_TOKEN]
//	    format: json
type Session struct {
	Host           string `json:"host"`
	Tenant         string `json:"tenant,omitempty"`
	Username       string `json:"username,omitempty"`
	Password       string `json:"password,omitempty"`
	Token          string `json:"token,omitempty"`
	TOTP           string `json:"totp,omitempty"`
	Mode           string `json:"mode,omitempty"`
	LoginType      string `json:"loginType,omitempty"`
	Description    string `json:"description,omitempty"`
	Certificate    string `json:"certificate,omitempty"`
	CertificateKey string `json:"certificateKey,omitempty"`
	Authorized     *bool  `json:"authorized,omitempty"`

	// SessionUri identifies where the session comes from, e.g. "bitwarden://<item-id>".
	// It is shown by c8y as the source of the session
	SessionUri string `json:"sessionUri,omitempty"`
}

// PrintSession writes the selected session to stdout for c8y. Session providers must not write
// anything else to stdout. Use stderr (and stdin) to interact with the user
func PrintSession(s Session) error {
	return FprintSession(os.Stdout, s)
}

// FprintSession writes the session as json to w
func FprintSession(w io.Writer, s Session) error {
	return Fprint(w, s)
}
