package extsdk

import (
	"encoding/json"
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

// SessionSummary is a session which a session provider lists, so c8y can show the provider's
// sessions in its own pickers (e.g. the browser session picker of "c8y mcp serve"). It must not
// contain any secrets. c8y passes the SessionUri to the provider's select command to get the
// session (with its secrets).
//
// Declare the list and select commands in extension.yaml:
//
//	sessionProviders:
//	  - name: myprovider
//	    exec: [sessions, select]       # interactive
//	    list: [sessions, list, --json] # prints the sessions with PrintSessions
//	    select: [sessions, get]        # followed by the sessionUri, prints the session with PrintSession
//
// Both commands are run without stdin, so they must not prompt the user.
type SessionSummary struct {
	// SessionUri identifies the session, e.g. "bitwarden://<item-id>" (required)
	SessionUri string `json:"sessionUri"`
	// Name shown to the user
	Name string `json:"name,omitempty"`
	// Host of the Cumulocity tenant (required)
	Host      string `json:"host"`
	Tenant    string `json:"tenant,omitempty"`
	Username  string `json:"username,omitempty"`
	Mode      string `json:"mode,omitempty"`
	LoginType string `json:"loginType,omitempty"`
	// Group is shown to the user, e.g. the folder of the vault
	Group string `json:"group,omitempty"`
}

// PrintSessions writes the sessions to stdout as a json array for c8y
func PrintSessions(sessions []SessionSummary) error {
	return FprintSessions(os.Stdout, sessions)
}

// FprintSessions writes the sessions as a json array to w
func FprintSessions(w io.Writer, sessions []SessionSummary) error {
	if sessions == nil {
		sessions = []SessionSummary{}
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(sessions)
}
