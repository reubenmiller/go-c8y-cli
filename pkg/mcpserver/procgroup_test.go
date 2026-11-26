//go:build !windows

package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// waitFor polls cond until it is true or the (generous) upper bound is reached.
func waitFor(t *testing.T, bound time.Duration, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(bound)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out after %s waiting for %s", bound, what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestRunnerTimeoutStopsChildProcesses(t *testing.T) {
	// A fake c8y which starts a child process that keeps the output open, like the
	// ssh process started by "remoteaccess connect ssh". It publishes the child pid
	// once it has written its output, so the test does not depend on how quickly
	// the process starts on a loaded machine.
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "child.pid")
	script := filepath.Join(dir, "c8y")
	content := "#!/bin/sh\nsleep 30 &\necho started\necho $! > '" + pidFile + ".tmp'\nmv '" + pidFile + ".tmp' '" + pidFile + "'\nwait\n"
	if err := os.WriteFile(script, []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}
	r := &Runner{Executable: script, BaseEnv: []string{"PATH=/usr/bin:/bin"}}
	plan := &Plan{Segments: []Segment{{Args: []string{"remoteaccess", "connect", "ssh", "--", "journalctl", "-f"}}}}

	// The context stands in for the timeout: it is cancelled only once the command
	// is known to be running, with a generous upper bound.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cancelled := make(chan time.Time, 1)
	go func() {
		for i := 0; i < 6000; i++ {
			if _, err := os.Stat(pidFile); err == nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		cancelled <- time.Now()
		cancel()
	}()

	result, err := r.Run(ctx, plan, "", "mcp:test", 1000)
	end := time.Now()
	if elapsed := end.Sub(<-cancelled); elapsed > 10*time.Second {
		t.Errorf("command should stop shortly after it is cancelled, took %s", elapsed)
	}
	if err == nil {
		t.Errorf("expected a timeout error")
	}
	if result == nil || result.Stdout != "started\n" {
		t.Errorf("expected the output before the timeout, got %+v", result)
	}

	// The child process must really be gone, not just no longer waited for
	raw, readErr := os.ReadFile(pidFile)
	if readErr != nil {
		t.Fatalf("the fake command never started its child process: %v", readErr)
	}
	pid, convErr := strconv.Atoi(strings.TrimSpace(string(raw)))
	if convErr != nil {
		t.Fatal(convErr)
	}
	waitFor(t, 10*time.Second, "the child process to be stopped", func() bool {
		return syscall.Kill(pid, 0) != nil
	})
}
