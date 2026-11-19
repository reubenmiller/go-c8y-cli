package extsdk

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func testEnv(extra ...string) []string {
	return append([]string{
		"C8Y_EXTENSION_CONTEXT_VERSION=1",
		"C8Y_EXTENSION_NAME=hello",
		"C8Y_EXTENSION_DIR=/data/c8y-hello",
		"C8Y_BIN=/usr/bin/c8y",
		"C8Y_TENANT=t123",
		"C8Y_USER=user1",
		"C8Y_SETTINGS_MODE=dev",
		"C8Y_SETTINGS_DEFAULTS_DRY=true",
		"OTHER=value",
	}, extra...)
}

func TestFromEnvironment(t *testing.T) {
	ctx, err := FromEnvironment(testEnv("C8Y_HOST=https://example.com"))
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Name != "hello" || ctx.Dir != "/data/c8y-hello" || ctx.Bin != "/usr/bin/c8y" {
		t.Errorf("unexpected extension values: %+v", ctx)
	}
	if ctx.Host != "https://example.com" || ctx.Tenant != "t123" || ctx.Username != "user1" || ctx.Mode != "dev" {
		t.Errorf("unexpected session values: %+v", ctx)
	}
	if !ctx.DryRun() || ctx.Insecure || ctx.NullInput() {
		t.Errorf("unexpected settings: dry=%v insecure=%v nullInput=%v", ctx.DryRun(), ctx.Insecure, ctx.NullInput())
	}
	if ctx, err := FromEnvironment(testEnv("C8Y_EXTENSION_NULL_INPUT=true")); err != nil || !ctx.NullInput() {
		t.Errorf("expected null input. err=%v", err)
	}
	if !ctx.HasSession() {
		t.Error("expected a session")
	}
	if got := ctx.ExtensionFile("templates/x.jsonnet"); got != filepath.Join("/data/c8y-hello", "templates", "x.jsonnet") {
		t.Errorf("unexpected extension file: %s", got)
	}
}

func TestFromEnvironmentNotRunByC8y(t *testing.T) {
	_, err := FromEnvironment([]string{"PATH=/bin"})
	if !errors.Is(err, ErrNotRunByC8y) {
		t.Fatalf("expected ErrNotRunByC8y, got %v", err)
	}
	if !strings.Contains(err.Error(), "c8y <extension>") {
		t.Errorf("error should explain how to run the extension: %s", err)
	}
}

func TestFromEnvironmentNewerVersion(t *testing.T) {
	_, err := FromEnvironment([]string{"C8Y_EXTENSION_CONTEXT_VERSION=2"})
	if !errors.Is(err, ErrUnsupportedContextVersion) {
		t.Fatalf("expected ErrUnsupportedContextVersion, got %v", err)
	}
}

func TestClientWithoutSession(t *testing.T) {
	ctx, err := FromEnvironment(testEnv())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ctx.Client(); !errors.Is(err, ErrNoSession) {
		t.Fatalf("expected ErrNoSession, got %v", err)
	}
}

func TestClientAuthorization(t *testing.T) {
	cases := []struct {
		name string
		env  []string
		want string
	}{
		{"token", []string{"C8Y_TOKEN=abc", "C8Y_HEADER_AUTHORIZATION=Basic xyz"}, "Bearer abc"},
		{"authorization header", []string{"C8Y_HEADER_AUTHORIZATION=Basic xyz"}, "Basic xyz"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r.Header.Get("Authorization")
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"name":"t123","domainName":"example.com"}`))
			}))
			defer srv.Close()

			ctx, err := FromEnvironment(testEnv(append(tc.env, "C8Y_HOST="+srv.URL)...))
			if err != nil {
				t.Fatal(err)
			}
			client, err := ctx.Client()
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := client.Tenant.GetCurrentTenant(context.Background()); err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("unexpected authorization header: %q", got)
			}
		})
	}
}

func TestCheckErrorUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"security/Unauthorized","message":"Invalid credentials!"}`))
	}))
	defer srv.Close()

	ctx, err := FromEnvironment(testEnv("C8Y_HOST="+srv.URL, "C8Y_TOKEN=expired"))
	if err != nil {
		t.Fatal(err)
	}
	client, err := ctx.Client()
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = client.Tenant.GetCurrentTenant(context.Background())
	if err == nil {
		t.Fatal("expected an error")
	}
	err = CheckError(err)
	if !strings.Contains(err.Error(), "c8y sessions login") {
		t.Errorf("error should tell the user to log in again: %s", err)
	}

	other := errors.New("other")
	if CheckError(other) != other {
		t.Error("other errors should be returned unchanged")
	}
}

func TestFprint(t *testing.T) {
	type item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var buf bytes.Buffer
	if err := Fprint(&buf, []item{{"1", "a"}, {"2", "<b>"}}); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "{\"id\":\"1\",\"name\":\"a\"}\n{\"id\":\"2\",\"name\":\"<b>\"}\n" {
		t.Errorf("unexpected output: %q", got)
	}

	buf.Reset()
	if err := Fprint(&buf, item{"1", "a"}); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "{\"id\":\"1\",\"name\":\"a\"}\n" {
		t.Errorf("unexpected output: %q", got)
	}

	buf.Reset()
	if err := Fprint(&buf, map[string]any{"values": []int{1, 2}}); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "{\"values\":[1,2]}\n" {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestExec(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}
	bin := filepath.Join(t.TempDir(), "c8y")
	script := "#!/bin/sh\necho \"args=$* host=$C8Y_HOST\"\necho err >&2\nexit 3\n"
	if err := os.WriteFile(bin, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	ctx, err := FromEnvironment(testEnv("C8Y_BIN="+bin, "C8Y_HOST=https://example.com"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := Exec(ctx, "devices", "get", "--id", "123")
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected an exit error, got %v", err)
	}
	if result.ExitCode != 3 {
		t.Errorf("unexpected exit code: %d", result.ExitCode)
	}
	if got := strings.TrimSpace(string(result.Stdout)); got != "args=devices get --id 123 host=https://example.com" {
		t.Errorf("unexpected stdout: %q", got)
	}
	if strings.TrimSpace(string(result.Stderr)) != "err" {
		t.Errorf("unexpected stderr: %q", result.Stderr)
	}

	if _, err := Exec(&Context{}, "devices"); !errors.Is(err, ErrNoC8yBin) {
		t.Errorf("expected ErrNoC8yBin, got %v", err)
	}
}

func TestNewRootCommandCompletion(t *testing.T) {
	root := NewRootCommand("c8y-hello")
	root.AddCommand(&cobra.Command{Use: "services", Short: "Manage services", Run: func(*cobra.Command, []string) {}})

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{cobra.ShellCompRequestCmd, "ser"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "services\tManage services") {
		t.Errorf("unexpected completion: %s", out.String())
	}

	out.Reset()
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "completion") {
		t.Errorf("completion command should be hidden: %s", out.String())
	}
}

func TestFprintSession(t *testing.T) {
	var buf bytes.Buffer
	authorized := true
	err := FprintSession(&buf, Session{
		Host:       "https://example.com",
		Tenant:     "t123",
		Username:   "user1",
		Token:      "abc",
		Authorized: &authorized,
		SessionUri: "vault://item/1",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"host":"https://example.com","tenant":"t123","username":"user1","token":"abc","authorized":true,"sessionUri":"vault://item/1"}` + "\n"
	if buf.String() != want {
		t.Errorf("unexpected session json:\n%s\nwant:\n%s", buf.String(), want)
	}
}
