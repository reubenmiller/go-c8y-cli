package fileutilities

import (
	"os"
	"os/user"
	"runtime"
	"strconv"
)

// CreateDirs create directory. All non-existing nested paths will be created.
// The folder owner will also be changed to match the current user (on non-windows systems only)
func CreateDirs(p string) error {
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		return err
	}

	// Change file ownership
	if runtime.GOOS != "windows" {
		var uid, gid int
		if os.Geteuid() == 0 {
			currentUser, err := user.Lookup(os.Getenv("SUDO_USER"))

			if err != nil {
				return err
			}

			uid, _ = strconv.Atoi(currentUser.Uid)
			gid, _ = strconv.Atoi(currentUser.Gid)

		} else {
			uid = os.Getuid()
			gid = os.Getgid()
		}
		if err := os.Chown(p, uid, gid); err != nil {
			return err
		}
	}
	return nil
}
