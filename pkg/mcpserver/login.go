package mcpserver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

// InteractiveLogin logs in once using the existing "c8y sessions login" command
// (so the user can enter the passphrase, 2FA code or complete an SSO login in
// the terminal), then provides the resulting token to the commands. The login
// is repeated in the terminal when the token is about to expire.
type InteractiveLogin struct {
	// Executable path to the c8y binary
	Executable string

	// SessionFile session file to log in with. If empty, the session is read
	// from environment variables
	SessionFile string

	// SessionMode mode passed to the login command (to avoid the mode prompt)
	SessionMode string

	// Stdin, Stderr terminal used for the login prompts
	Stdin  io.Reader
	Stderr io.Writer

	// Terminal lock which is held while the login uses the terminal (shared with
	// the connection approval prompt)
	Terminal sync.Locker

	// RenewBefore log in again when the token expires within this duration
	RenewBefore time.Duration

	// TempDir directory for the copy of the session file without secrets
	TempDir string

	// NoTerminal the server has no terminal (stdio mode), so the session can only be
	// unlocked without user input or with the pin entry command (settings.pinEntry)
	NoTerminal bool

	mu        sync.Mutex
	env       []string
	expiresAt *time.Time
	usesToken bool
	loggedIn  bool
}

// Login runs the interactive login
func (l *InteractiveLogin) Login(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.login(ctx)
}

func (l *InteractiveLogin) login(ctx context.Context) error {
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
	cmd.Stdin = l.Stdin
	cmd.Stderr = l.Stderr
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	if l.Terminal != nil {
		l.Terminal.Lock()
	}
	err := cmd.Run()
	if l.Terminal != nil {
		l.Terminal.Unlock()
	}
	if err != nil {
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
		if err := l.login(ctx); err != nil {
			if l.NoTerminal {
				return nil, fmt.Errorf("could not log in to the Cumulocity session. Without a terminal, an encrypted session is unlocked with the pin entry command (settings.pinEntry, e.g. a password manager or keychain helper) or C8Y_PASSPHRASE, and two-factor authentication needs a TOTP secret in the session. Ask the user to check this, or to run \"c8y mcp serve\" in a terminal and register \"c8y mcp connect\" instead. %w", err)
			}
			return nil, fmt.Errorf("the Cumulocity session expired and the login in the 'c8y mcp serve' terminal failed. Ask the user to log in again in that terminal. %w", err)
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
