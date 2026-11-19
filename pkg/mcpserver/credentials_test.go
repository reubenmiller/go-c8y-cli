//go:build !windows

package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type staticAuth []string

func (a staticAuth) AuthEnv(ctx context.Context) ([]string, error) { return a, nil }

func TestRunnerPassesCredentialsViaPipe(t *testing.T) {
	// A fake c8y which prints its environment and what it can read from the credentials fd
	script := filepath.Join(t.TempDir(), "c8y")
	content := "#!/bin/sh\necho \"env: $(env | grep '^C8Y_' | sort | tr '\\n' ' ')\"\necho \"fd: $(cat <&${C8Y_CREDENTIALS_FD:-3})\"\n"
	if err := os.WriteFile(script, []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}

	r := &Runner{
		Executable: script,
		BaseEnv:    []string{"PATH=/usr/bin:/bin", "C8Y_TOKEN=inherited", "C8Y_PASSWORD=inherited"},
		Auth:       staticAuth{"C8Y_HOST=https://example.com", "C8Y_TOKEN=secret-token", "C8Y_PASSWORD=secret-pass"},
	}
	// two segments, so each command needs its own pipe
	plan := &Plan{Segments: []Segment{{Args: []string{"devices", "list"}, Read: true}, {Args: []string{"util", "show"}, Read: true}}}
	result, err := r.Run(context.Background(), plan, "", "mcp:test", 10_000)
	if err != nil {
		t.Fatalf("unexpected error: %s (stderr: %s)", err, result.Stderr)
	}

	env, fd, _ := strings.Cut(result.Stdout, "fd: ")
	if strings.Contains(env, "secret") || strings.Contains(env, "inherited") {
		t.Errorf("credentials should not be in the environment: %s", env)
	}
	if !strings.Contains(env, "C8Y_HOST=https://example.com") || !strings.Contains(env, "C8Y_CREDENTIALS_FD=3") {
		t.Errorf("expected the host and the credentials fd in the environment: %s", env)
	}
	if !strings.Contains(fd, `"token":"secret-token"`) || !strings.Contains(fd, `"password":"secret-pass"`) {
		t.Errorf("expected the credentials on the fd: %s", fd)
	}
}

func TestRunnerAddsBatchModeToDeviceShellCommands(t *testing.T) {
	// A fake c8y which prints its arguments
	script := filepath.Join(t.TempDir(), "c8y")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho \"$@\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	r := &Runner{Executable: script, BaseEnv: []string{"PATH=/usr/bin:/bin"}, SSHFallbackUser: "root"}
	plan := &Plan{Segments: []Segment{{Args: []string{"remoteaccess", "connect", "ssh", "--device", "1", "--batch=false", "--", "uptime"}, CommandPath: "remoteaccess connect ssh", DeviceShell: true}}}
	result, err := r.Run(context.Background(), plan, "", "mcp:test", 10_000)
	if err != nil {
		t.Fatalf("unexpected error: %s (stderr: %s)", err, result.Stderr)
	}
	// the injected flag comes after the agent's flags (so it wins), and before "--"
	if got := strings.TrimSpace(result.Stdout); !strings.HasSuffix(got, "--batch=false --logMessage mcp:test --force --batch --fallback-user root -- uptime") {
		t.Errorf("expected --batch before the argument terminator, got %q", got)
	}

	plan = &Plan{Segments: []Segment{{Args: []string{"devices", "list"}, Read: true}}}
	result, _ = r.Run(context.Background(), plan, "", "mcp:test", 10_000)
	if strings.Contains(result.Stdout, "--batch") || strings.Contains(result.Stdout, "--fallback-user") {
		t.Errorf("expected no --batch for other commands, got %q", result.Stdout)
	}

	// A c8y_Command operation is a device shell command, but not an ssh one: the ssh
	// flags would make "operations create" and "api" fail with "unknown flag"
	for _, seg := range []Segment{
		{Args: []string{"operations", "create", "--device", "1", "--data", "c8y_Command.text=uname"}, CommandPath: "operations create", DeviceShell: true},
		{Args: []string{"api", "POST", "/devicecontrol/operations", "--data", "c8y_Command.text=uname"}, CommandPath: "api", DeviceShell: true},
	} {
		plan = &Plan{Segments: []Segment{seg}}
		result, err = r.Run(context.Background(), plan, "", "mcp:test", 10_000)
		if err != nil {
			t.Fatalf("unexpected error: %s (stderr: %s)", err, result.Stderr)
		}
		got := strings.TrimSpace(result.Stdout)
		if strings.Contains(got, "--batch") || strings.Contains(got, "--fallback-user") {
			t.Errorf("expected no ssh flags for %q, got %q", seg.CommandPath, got)
		}
		if !strings.HasSuffix(got, "--logMessage mcp:test --force") {
			t.Errorf("expected the write flags for %q, got %q", seg.CommandPath, got)
		}
	}
}

func TestRunnerAddsFlagsBeforeExtensionCommands(t *testing.T) {
	// A fake c8y which prints its arguments
	script := filepath.Join(t.TempDir(), "c8y")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho \"$@\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	r := &Runner{Executable: script, BaseEnv: []string{"PATH=/usr/bin:/bin"}, SSHFallbackUser: "root"}

	cases := []struct {
		seg  Segment
		want string
	}{
		{
			// c8y parses the flags before the command name, and does not pass them to the extension
			seg:  Segment{Args: []string{"tedge", "demo", "start", "tedge_mux_test", "--skip-website"}, CommandPath: "tedge demo start", Extension: true},
			want: "--logMessage mcp:test --force tedge demo start tedge_mux_test --skip-website",
		},
		{
			seg:  Segment{Args: []string{"tedge", "demo", "list"}, CommandPath: "tedge demo list", Extension: true, Read: true},
			want: "--logMessage mcp:test tedge demo list",
		},
		{
			// built-in commands are unchanged
			seg:  Segment{Args: []string{"devices", "delete", "--id", "1"}, CommandPath: "devices delete"},
			want: "devices delete --id 1 --logMessage mcp:test --force",
		},
	}
	for _, tc := range cases {
		result, err := r.Run(context.Background(), &Plan{Segments: []Segment{tc.seg}}, "", "mcp:test", 10_000)
		if err != nil {
			t.Fatalf("unexpected error: %s (stderr: %s)", err, result.Stderr)
		}
		if got := strings.TrimSpace(result.Stdout); got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.seg.CommandPath, got, tc.want)
		}
	}
}
