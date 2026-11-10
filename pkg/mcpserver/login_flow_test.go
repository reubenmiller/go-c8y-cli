package mcpserver

import (
	"context"
	"encoding/json"
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

func shortTempDir(t *testing.T) string {
	t.Helper()
	// t.TempDir() can exceed the unix socket path limit on macOS
	dir, err := os.MkdirTemp("/tmp", "c8ymcp")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestParseDotEnv(t *testing.T) {
	values := parseDotEnv("C8Y_HOST=https://example.com\n\n# comment\nC8Y_TOKEN=abc=def\nINVALID\n")
	if values["C8Y_HOST"] != "https://example.com" || values["C8Y_TOKEN"] != "abc=def" || len(values) != 2 {
		t.Errorf("unexpected values: %v", values)
	}
}

func TestTokenExpiresAt(t *testing.T) {
	expires := time.Now().Add(time.Hour).Truncate(time.Second)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expires)}).SignedString([]byte("key"))
	if err != nil {
		t.Fatal(err)
	}
	got := TokenExpiresAt(token)
	if got == nil || !got.Equal(expires) {
		t.Errorf("unexpected expiry %v", got)
	}
	if TokenExpiresAt("not-a-token") != nil || TokenExpiresAt("") != nil {
		t.Errorf("invalid tokens should not have an expiry")
	}
}

func TestRunnerEnvWithAuthProvider(t *testing.T) {
	r := &Runner{
		SessionFile: "/original/session.json",
		BaseEnv: []string{
			"PATH=/usr/bin",
			"C8Y_SESSION=/original/session.json",
			"C8Y_PASSWORD=secret",
			"C8Y_PASSPHRASE=secret",
			"C8Y_TOKEN=old",
			"C8Y_SESSION_HOME=/home/user/.cumulocity",
		},
	}
	env := strings.Join(r.Env(config.SessionModeQual, []string{"C8Y_TOKEN=new", "C8Y_SESSION=/tmp/copy.json"}), "\n")
	for _, unwanted := range []string{"C8Y_PASSWORD", "C8Y_PASSPHRASE", "C8Y_TOKEN=old", "/original/session.json"} {
		if strings.Contains(env, unwanted) {
			t.Errorf("env should not contain %q", unwanted)
		}
	}
	for _, wanted := range []string{"PATH=/usr/bin", "C8Y_SESSION_HOME=/home/user/.cumulocity", "C8Y_TOKEN=new", "C8Y_SESSION=/tmp/copy.json", "C8Y_MODE=qual"} {
		if !strings.Contains(env, wanted) {
			t.Errorf("env should contain %q", wanted)
		}
	}
}

// writeFakeC8y creates a script which mimics "c8y sessions login --output-format dotenv"
func writeFakeC8y(t *testing.T, dir string, token string) string {
	t.Helper()
	script := filepath.Join(dir, "c8y")
	content := fmt.Sprintf("#!/bin/sh\necho \"login prompt\" >&2\necho \"args: $*\" >> %s/args.log\n"+
		"printf 'C8Y_HOST=https://example.com\\nC8Y_TENANT=t123\\nC8Y_USER=agent\\nC8Y_TOKEN=%s\\nC8Y_PASSWORD=secret\\nC8Y_PASSPHRASE=secret\\nC8Y_HEADER_AUTHORIZATION=Bearer x\\n'\n",
		dir, token)
	if err := os.WriteFile(script, []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}
	return script
}

func TestInteractiveLogin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}
	dir := shortTempDir(t)

	expires := time.Now().Add(time.Hour)
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expires)}).SignedString([]byte("key"))

	sessionFile := filepath.Join(dir, "my-session.json")
	session := `{"host":"https://example.com","username":"agent","password":"{encrypted}abc","token":"{encrypted}def","totp":"x","settings":{"session":{"mode":"dev"},"defaults":{"pageSize":50},"other":{"secret":"{encrypted}zzz"}}}`
	if err := os.WriteFile(sessionFile, []byte(session), 0o600); err != nil {
		t.Fatal(err)
	}

	login := &InteractiveLogin{
		Executable:  writeFakeC8y(t, dir, token),
		SessionFile: sessionFile,
		SessionMode: "dev",
		Stderr:      &strings.Builder{},
		TempDir:     dir, // same directory as the session file, which must not be overwritten
	}
	env, err := login.AuthEnv(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(env, "\n")
	for _, unwanted := range []string{"C8Y_PASSWORD", "C8Y_PASSPHRASE", "C8Y_HEADER"} {
		if strings.Contains(joined, unwanted) {
			t.Errorf("env should not contain %q when a token is available", unwanted)
		}
	}
	for _, wanted := range []string{"C8Y_HOST=https://example.com", "C8Y_TENANT=t123", "C8Y_TOKEN=" + token} {
		if !strings.Contains(joined, wanted) {
			t.Errorf("env should contain %q", wanted)
		}
	}

	args, _ := os.ReadFile(filepath.Join(dir, "args.log"))
	if !strings.Contains(string(args), "sessions login --output-format dotenv --no-banner --from-file "+sessionFile+" --mode dev") {
		t.Errorf("unexpected login args: %s", args)
	}

	// The commands use a copy of the session file without secrets
	var copyPath string
	for _, item := range env {
		if v, ok := strings.CutPrefix(item, "C8Y_SESSION="); ok {
			copyPath = v
		}
	}
	if copyPath == "" || copyPath == sessionFile {
		t.Fatalf("expected a copy of the session file, got %q", copyPath)
	}
	if original, _ := os.ReadFile(sessionFile); string(original) != session {
		t.Fatalf("the original session file was modified: %s", original)
	}
	info, err := os.Stat(copyPath)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Errorf("unexpected permissions %v", info.Mode().Perm())
	}
	contents, _ := os.ReadFile(copyPath)
	for _, unwanted := range []string{"encrypted", "password", "token", "totp"} {
		if strings.Contains(string(contents), unwanted) {
			t.Errorf("session copy should not contain %q: %s", unwanted, contents)
		}
	}
	data := map[string]any{}
	if err := json.Unmarshal(contents, &data); err != nil {
		t.Fatal(err)
	}
	if data["host"] != "https://example.com" {
		t.Errorf("session copy should keep the host: %s", contents)
	}

	// The token is reused while it is valid
	if _, err := login.AuthEnv(context.Background()); err != nil {
		t.Fatal(err)
	}
	args, _ = os.ReadFile(filepath.Join(dir, "args.log"))
	if strings.Count(string(args), "sessions login") != 1 {
		t.Errorf("expected a single login, got: %s", args)
	}
}

func TestInteractiveLoginRenewsExpiringToken(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}
	dir := shortTempDir(t)
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Second))}).SignedString([]byte("key"))

	login := &InteractiveLogin{
		Executable: writeFakeC8y(t, dir, token),
		Stderr:     &strings.Builder{},
		TempDir:    dir,
	}
	for range 2 {
		if _, err := login.AuthEnv(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	args, _ := os.ReadFile(filepath.Join(dir, "args.log"))
	if strings.Count(string(args), "sessions login") != 2 {
		t.Errorf("expected the login to be repeated for an expiring token, got: %s", args)
	}
	if !strings.Contains(string(args), "--from-env") {
		t.Errorf("expected --from-env when no session file is used: %s", args)
	}
}

func newTestServer() *Server {
	policy := DefaultPolicy()
	return &Server{
		Version:    "test",
		Session:    SessionInfo{Host: "https://example.com", Tenant: "t123", SessionMode: "dev"},
		Policy:     policy,
		Classifier: &Classifier{Root: newTestRoot(), Policy: policy},
		Runner:     &Runner{Executable: "/bin/false", Mode: config.SessionModeProduction},
	}
}
