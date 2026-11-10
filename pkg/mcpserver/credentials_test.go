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
