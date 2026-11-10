package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

// newFakeLogin returns a login which runs a fake "c8y sessions login". It records the
// passphrase it receives via the credentials pipe, and fails without one
func newFakeLogin(t *testing.T, token string) (*InteractiveLogin, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}
	dir := t.TempDir()
	record := filepath.Join(dir, "calls")
	script := filepath.Join(dir, "c8y")
	content := fmt.Sprintf(`#!/bin/sh
passphrase=""
if [ -n "$C8Y_CREDENTIALS_FD" ]; then
  passphrase=$(cat <&"$C8Y_CREDENTIALS_FD")
fi
in_env=$(env | grep -c secret-pass)
echo "passphrase=$passphrase env=$in_env" >> %q
case "$passphrase" in
  *secret-pass*) ;;
  *) printf '\033[31mERROR\033[0m could not decrypt the session\n\n' >&2; exit 1;;
esac
case "$passphrase" in
  *need-totp*)
    case "$passphrase" in
      *'"totpCode":"123456"'*) ;;
      *) printf 'ERROR %%s\n' "%s" >&2; exit 1;;
    esac;;
esac
echo "C8Y_HOST=https://example.com"
echo "C8Y_TOKEN=%s"
`, record, config.TOTPCodeRequiredMessage, token)
	if err := os.WriteFile(script, []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}
	return &InteractiveLogin{Executable: script, TempDir: dir}, record
}

func readCalls(t *testing.T, record string) []string {
	t.Helper()
	b, _ := os.ReadFile(record)
	return strings.Fields(strings.ReplaceAll(string(b), "\n", " "))
}

func testToken(t *testing.T, expires time.Time) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expires)}).SignedString([]byte("k"))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestLoginWithPassphrase(t *testing.T) {
	ctx := context.Background()
	login, record := newFakeLogin(t, testToken(t, time.Now().Add(time.Hour)))

	err := login.Login(ctx)
	if err == nil || !strings.Contains(err.Error(), "ERROR could not decrypt the session") || strings.Contains(err.Error(), "\033") {
		t.Errorf("expected the reason (without colors) in the error, got: %v", err)
	}
	if err := login.LoginWith(ctx, LoginSecrets{Passphrase: "secret-pass"}); err != nil {
		t.Fatal(err)
	}
	calls := readCalls(t, record)
	// the passphrase is passed via the pipe, not the environment
	want := []string{"passphrase=", "env=0", `passphrase={"passphrase":"secret-pass"}`, "env=0"}
	if strings.Join(calls, " ") != strings.Join(want, " ") {
		t.Errorf("unexpected calls: %q", calls)
	}
	env, err := login.AuthEnv(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(env, " "), "secret-pass") {
		t.Errorf("the passphrase must not be passed to the commands: %q", env)
	}
}

func TestLoginRenew(t *testing.T) {
	ctx := context.Background()
	// the token expires within the renewal period, so each command logs in again
	login, record := newFakeLogin(t, testToken(t, time.Now().Add(time.Minute)))
	renewals := 0
	login.Renew = func(ctx context.Context, login func(context.Context, LoginSecrets) error) error {
		renewals++
		// e.g. the user first enters a wrong passphrase in the browser
		if err := login(ctx, LoginSecrets{Passphrase: "wrong"}); err == nil {
			t.Errorf("expected the wrong passphrase to fail")
		}
		return login(ctx, LoginSecrets{Passphrase: "secret-pass"})
	}
	if err := login.LoginWith(ctx, LoginSecrets{Passphrase: "secret-pass"}); err != nil {
		t.Fatal(err)
	}
	if _, err := login.AuthEnv(ctx); err != nil {
		t.Fatal(err)
	}
	if renewals != 1 {
		t.Errorf("expected the login to be renewed via Renew, got %d renewals", renewals)
	}
	if calls := readCalls(t, record); len(calls) != 6 {
		t.Errorf("expected 3 logins, got %q", calls)
	}

	// a failed renewal is reported to the agent
	login.Renew = func(ctx context.Context, login func(context.Context, LoginSecrets) error) error {
		return ErrPickerCancelled
	}
	if _, err := login.AuthEnv(ctx); err == nil || !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("expected the renewal error, got %v", err)
	}
}

func TestLoginTOTP(t *testing.T) {
	ctx := context.Background()
	login, record := newFakeLogin(t, testToken(t, time.Now().Add(time.Hour)))

	// the login reports that a code is needed, instead of prompting
	err := login.LoginWith(ctx, LoginSecrets{Passphrase: "secret-pass need-totp"})
	if !errors.Is(err, ErrTOTPRequired) {
		t.Fatalf("expected ErrTOTPRequired, got %v", err)
	}
	if err := login.LoginWith(ctx, LoginSecrets{Passphrase: "secret-pass need-totp", TOTPCode: "123456"}); err != nil {
		t.Fatal(err)
	}
	calls := readCalls(t, record)
	if last := calls[len(calls)-2]; !strings.Contains(last, `"totpCode":"123456"`) {
		t.Errorf("the code should be passed via the pipe: %q", calls)
	}
}
