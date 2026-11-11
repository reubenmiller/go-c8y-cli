//go:build !windows

package mcpserver

import (
	"os/exec"
	"syscall"
)

// setProcessGroup starts the command in its own process group, and kills the whole
// group when the command is cancelled (e.g. on timeout)
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		// a negative pid signals the process group
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
