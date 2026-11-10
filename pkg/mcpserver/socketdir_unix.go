//go:build !windows

package mcpserver

import (
	"fmt"
	"os"
	"syscall"
)

// CheckSocketDir checks that the directory containing the socket can only be
// modified by the current user. Otherwise another user could replace the socket
// with their own, and the MCP client would connect to their server instead.
func CheckSocketDir(dir string) error {
	// Follow symlinks: the target directory is the one which needs to be protected
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("could not check socket directory. %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("socket directory is not a directory: %s", dir)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("could not read the owner of the socket directory: %s", dir)
	}
	if int(stat.Uid) != os.Getuid() {
		return fmt.Errorf("socket directory must be owned by the current user (uid %d), but is owned by uid %d: %s", os.Getuid(), stat.Uid, dir)
	}
	if info.Mode().Perm()&0o022 != 0 {
		return fmt.Errorf("socket directory must not be writable by other users (mode %04o). Use a private directory, e.g. chmod go-w %s", info.Mode().Perm(), dir)
	}
	return nil
}
