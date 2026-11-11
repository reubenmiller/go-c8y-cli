//go:build windows

package mcpserver

import "os/exec"

// setProcessGroup is not supported on Windows. Only the c8y process is stopped when
// the command is cancelled, and WaitDelay stops waiting for its output.
func setProcessGroup(cmd *exec.Cmd) {}
