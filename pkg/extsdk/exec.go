package extsdk

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
)

// ErrNoC8yBin is returned by Exec when the path to c8y is unknown
var ErrNoC8yBin = errors.New("the path to the c8y executable is unknown (C8Y_BIN is not set)")

// ExecResult is the result of running a c8y command
type ExecResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// Exec runs a c8y command with the same c8y executable and environment (session) which
// started the extension, e.g. Exec(ctx, "devices", "get", "--id", "123").
// A non-zero exit code is returned in the result and as an *exec.ExitError
func Exec(c *Context, args ...string) (*ExecResult, error) {
	return ExecContext(context.Background(), c, args...)
}

// ExecContext is like Exec but uses a context to cancel the command
func ExecContext(ctx context.Context, c *Context, args ...string) (*ExecResult, error) {
	if c.Bin == "" {
		return nil, ErrNoC8yBin
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, c.Bin, args...)
	cmd.Env = c.Environ()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	result := &ExecResult{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
	}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	return result, err
}
