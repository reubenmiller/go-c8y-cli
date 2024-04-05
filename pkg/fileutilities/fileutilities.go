package fileutilities

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
)

// CreateDirs create directory. All non-existing nested paths will be created.
// The folder owner will also be changed to match the current user (on non-windows systems only)
func CreateDirs(p string) error {
	if p == "" {
		return nil
	}
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		return err
	}

	// Change file ownership
	if runtime.GOOS != "windows" {
		owner := os.Getenv("SUDO_USER")

		if os.Geteuid() != 0 || owner == "" {
			// Can change ownership if not running as sudo
			// or user is root, and root should not take over ownership
			return nil
		}

		if os.Geteuid() == 0 && owner == "" {
			// TODO: Check if the user has access to it, if so don't change ownership
			// Is root user: don't try changing permissions to the root user
			return nil
		}

		// Note: os.Chown can't be used as os/user.Lookup is not reliable on macOS
		// golang bug: https://github.com/golang/go/issues/24383
		cmd := exec.Command("chown", "-R", "-L", owner, p)

		b, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error change owner of dir %s to %s: %w %s", p, owner, err, b)
		}
	}
	return nil
}

var ErrInvalid = errors.New("invalid file destination")

// Copy copies src to dst like the cp command.
func CopyFile(dst, src string) error {
	if dst == src {
		return ErrInvalid
	}

	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	info, err := srcF.Stat()
	if err != nil {
		return err
	}

	dstF, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dstF.Close()

	if _, err := io.Copy(dstF, srcF); err != nil {
		return err
	}
	return nil
}
