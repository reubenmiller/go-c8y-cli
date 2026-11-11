//go:build !windows

package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunnerTimeoutStopsChildProcesses(t *testing.T) {
	// A fake c8y which starts a child process that keeps the output open, like the
	// ssh process started by "remoteaccess connect ssh"
	script := filepath.Join(t.TempDir(), "c8y")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 30 &\necho started\nwait\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	r := &Runner{Executable: script, BaseEnv: []string{"PATH=/usr/bin:/bin"}}
	plan := &Plan{Segments: []Segment{{Args: []string{"remoteaccess", "connect", "ssh", "--", "journalctl", "-f"}}}}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	start := time.Now()
	result, err := r.Run(ctx, plan, "", "mcp:test", 1000)
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("command should stop shortly after the timeout, took %s", elapsed)
	}
	if err == nil {
		t.Errorf("expected a timeout error")
	}
	if result == nil || result.Stdout != "started\n" {
		t.Errorf("expected the output before the timeout, got %+v", result)
	}
}
