//go:build windows

package mcpserver

// CheckSocketDir is not supported on Windows, where the socket is protected by
// the ACLs of the user's profile directory
func CheckSocketDir(dir string) error {
	return nil
}
