package mcpserver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/spf13/viper"
)

// childAuthEnvVars credential related variables which are passed from the login
// result to the commands run for the agent. Secrets such as the password and
// the encryption passphrase are not included unless they are needed.
var childAuthEnvVars = []string{
	"C8Y_HOST",
	"C8Y_URL",
	"C8Y_BASEURL",
	"C8Y_DOMAIN",
	"C8Y_TENANT",
	"C8Y_VERSION",
	"C8Y_USER",
	"C8Y_USERNAME",
	"C8Y_TOKEN",
	"C8Y_SETTINGS_LOGIN_TYPE",
}

// inheritedAuthEnvVars credential related variables which are removed from the
// server's environment before running commands, so the commands only use the
// credentials provided by the login
var inheritedAuthEnvVars = []string{
	"C8Y_SESSION",
	"C8Y_HOST",
	"C8Y_URL",
	"C8Y_BASEURL",
	"C8Y_DOMAIN",
	"C8Y_TENANT",
	"C8Y_VERSION",
	"C8Y_USER",
	"C8Y_USERNAME",
	"C8Y_PASSWORD",
	"C8Y_TOKEN",
	"C8Y_TOTP",
	"C8Y_HEADER",
	"C8Y_HEADER_AUTHORIZATION",
	"C8Y_SETTINGS_LOGIN_TYPE",
	"C8Y_CERTIFICATE",
	"C8Y_CERTIFICATE_KEY",
	"C8Y_PASSPHRASE",
	"C8Y_PASSPHRASE_TEXT",
}

// sessionSecretKeys session file properties which are removed from the copy of
// the session file used by the commands run for the agent
var sessionSecretKeys = []string{"password", "token", "totp", "certificate", "certificatekey"}

// AuthProvider provides the credentials for the commands run for the agent
type AuthProvider interface {
	// AuthEnv returns the environment variables containing the credentials
	AuthEnv(ctx context.Context) ([]string, error)
}

// LoginSecrets secrets entered by the user for a login. They are passed to the login
// via an inherited pipe, and aren't kept
type LoginSecrets struct {
	// Passphrase to unlock an encrypted session
	Passphrase string
	// TOTPCode two-factor authentication code
	TOTPCode string
}

// ErrTOTPRequired the login needs a two-factor authentication code
var ErrTOTPRequired = errors.New(config.TOTPCodeRequiredMessage)

// InteractiveLogin logs in using the existing "c8y sessions login" command, then
// provides the resulting token to the commands. The server has no terminal, so the
// login can't prompt: an encrypted session is unlocked with the pin entry command
// (settings.pinEntry) or the passphrase given in LoginSecrets, and a two-factor
// code is generated from the session's TOTP secret or given in LoginSecrets. The
// login is repeated when the token is about to expire (see Renew).
type InteractiveLogin struct {
	// Executable path to the c8y binary
	Executable string

	// SessionFile session file to log in with. If empty, the session is read
	// from environment variables
	SessionFile string

	// SessionMode mode passed to the login command (to avoid the mode prompt)
	SessionMode string

	// Stderr receives the login's messages
	Stderr io.Writer

	// RenewBefore log in again when the token expires within this duration
	RenewBefore time.Duration

	// TempDir directory for the copy of the session file without secrets
	TempDir string

	// Renew optional function which logs in again when the token expires, instead of
	// logging in without secrets. It is used to ask the user for the secrets in the
	// browser. It must call login (which may be called several times, e.g. after a wrong
	// passphrase) and return its last error. It is called while the login is locked
	Renew func(ctx context.Context, login func(ctx context.Context, secrets LoginSecrets) error) error

	mu        sync.Mutex
	env       []string
	expiresAt *time.Time
	usesToken bool
	loggedIn  bool
}

// Login runs the interactive login
func (l *InteractiveLogin) Login(ctx context.Context) error {
	return l.LoginWith(ctx, LoginSecrets{})
}

// LoginWith runs the login with the secrets entered by the user (if any). It returns
// ErrTOTPRequired if a two-factor code is needed
func (l *InteractiveLogin) LoginWith(ctx context.Context, secrets LoginSecrets) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.login(ctx, secrets)
}

func (l *InteractiveLogin) login(ctx context.Context, secrets LoginSecrets) error {
	args := []string{"sessions", "login", "--output-format", "dotenv", "--no-banner"}
	if l.SessionFile != "" {
		args = append(args, "--from-file", l.SessionFile)
	} else {
		args = append(args, "--from-env")
	}
	if l.SessionMode != "" {
		args = append(args, "--mode", l.SessionMode)
	}

	cmd := exec.CommandContext(ctx, l.Executable, args...)
	// keep the end of the output, to explain why the login failed
	stderrTail := &tailWriter{max: 2048}
	if l.Stderr != nil {
		cmd.Stderr = io.MultiWriter(l.Stderr, stderrTail)
	} else {
		cmd.Stderr = stderrTail
	}
	if secrets != (LoginSecrets{}) {
		cmd.Env = os.Environ()
		if err := passCredentials(cmd, &config.Credentials{Passphrase: secrets.Passphrase, TOTPCode: secrets.TOTPCode}); err != nil {
			return err
		}
	}
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		reason := stderrTail.lastLine()
		if strings.Contains(reason, config.TOTPCodeRequiredMessage) {
			return ErrTOTPRequired
		}
		if reason != "" {
			return fmt.Errorf("login failed. %s", reason)
		}
		return fmt.Errorf("login failed. %w", err)
	}

	values := parseDotEnv(stdout.String())
	if values["C8Y_HOST"] == "" {
		return fmt.Errorf("login did not return a session")
	}

	env := make([]string, 0, len(childAuthEnvVars)+3)
	for _, key := range childAuthEnvVars {
		if v := values[key]; v != "" {
			env = append(env, key+"="+v)
		}
	}

	token := values["C8Y_TOKEN"]
	if token == "" {
		// Basic auth (or certificate auth without a token) needs the
		// credentials for every request
		for _, key := range []string{"C8Y_PASSWORD", "C8Y_CERTIFICATE", "C8Y_CERTIFICATE_KEY"} {
			if v := values[key]; v != "" {
				env = append(env, key+"="+v)
			}
		}
	}

	if l.SessionFile != "" {
		sanitized, err := l.writeSanitizedSession()
		if err != nil {
			return err
		}
		env = append(env, "C8Y_SESSION="+sanitized)
	}

	l.env = env
	l.expiresAt = TokenExpiresAt(token)
	l.usesToken = token != ""
	l.loggedIn = true
	return nil
}

// UsesToken returns true if the commands use a token (rather than basic auth)
func (l *InteractiveLogin) UsesToken() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.usesToken
}

// AuthEnv returns the credentials, logging in again if the token expires soon
func (l *InteractiveLogin) AuthEnv(ctx context.Context) ([]string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	renewBefore := l.RenewBefore
	if renewBefore == 0 {
		renewBefore = 2 * time.Minute
	}
	if !l.loggedIn || (l.expiresAt != nil && time.Now().Add(renewBefore).After(*l.expiresAt)) {
		if l.loggedIn && l.Stderr != nil {
			fmt.Fprintf(l.Stderr, "\nThe session token expires soon. Log in again to continue.\n")
		}
		var err error
		if l.Renew != nil {
			err = l.Renew(ctx, l.login)
		} else {
			err = l.login(ctx, LoginSecrets{})
		}
		if err != nil {
			return nil, fmt.Errorf("could not log in to the Cumulocity session. %w. Ask the user to check the session, or to switch to another session (switch_session prompt)", err)
		}
	}
	return append([]string{}, l.env...), nil
}

// ExpiresAt returns when the current token expires (if known)
func (l *InteractiveLogin) ExpiresAt() *time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.expiresAt
}

// writeSanitizedSession writes a copy of the session file without any secrets,
// so the commands keep the session's settings but can't access the credentials
func (l *InteractiveLogin) writeSanitizedSession() (string, error) {
	v := viper.New()
	v.SetConfigFile(l.SessionFile)
	if err := v.ReadInConfig(); err != nil {
		return "", fmt.Errorf("could not read session file. %w", err)
	}
	settings := v.AllSettings()
	for _, key := range sessionSecretKeys {
		delete(settings, key)
	}
	removeEncryptedValues(settings)

	b, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", err
	}

	dir := l.TempDir
	if dir == "" {
		dir = os.TempDir()
	}
	name := "mcp-session-" + strings.TrimSuffix(filepath.Base(l.SessionFile), filepath.Ext(l.SessionFile)) + ".json"
	path := filepath.Join(dir, name)
	if same, _ := filepath.Abs(path); same == l.SessionFile {
		return "", fmt.Errorf("refusing to overwrite the session file %s", l.SessionFile)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return "", fmt.Errorf("could not write session copy. %w", err)
	}
	return path, nil
}

func removeEncryptedValues(m map[string]any) {
	for k, v := range m {
		switch value := v.(type) {
		case string:
			if strings.HasPrefix(value, "{encrypted}") {
				delete(m, k)
			}
		case map[string]any:
			removeEncryptedValues(value)
		}
	}
}

func parseDotEnv(text string) map[string]string {
	values := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = value
	}
	return values
}

// TokenExpiresAt returns the expiry time of a JWT token, or nil if it can't be determined
func TokenExpiresAt(token string) *time.Time {
	if token == "" {
		return nil
	}
	claims := jwt.RegisteredClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(token, &claims); err != nil || claims.ExpiresAt == nil {
		return nil
	}
	t := claims.ExpiresAt.Time
	return &t
}

// tailWriter keeps the last bytes written to it
type tailWriter struct {
	max int
	mu  sync.Mutex
	buf []byte
}

func (w *tailWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf = append(w.buf, p...)
	if len(w.buf) > w.max {
		w.buf = w.buf[len(w.buf)-w.max:]
	}
	return len(p), nil
}

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)

// lastLine returns the last non-empty line, without colors
func (w *tailWriter) lastLine() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	lines := strings.Split(ansiEscape.ReplaceAllString(string(w.buf), ""), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			return line
		}
	}
	return ""
}
