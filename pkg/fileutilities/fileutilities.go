package fileutilities

import (
	"os"
	"runtime"
)

// CreateDirs create directory. All non-existing nested paths will be created.
// The folder owner will also be changed to match the current user (on non-windows systems only)
func CreateDirs(p string) error {
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		return err
	}

	// Change file ownership
	if runtime.GOOS != "windows" {
		if err := os.Chown(p, os.Getuid(), os.Getgid()); err != nil {
			return err
		}
	}
	return nil
}
