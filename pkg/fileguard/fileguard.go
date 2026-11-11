// Package fileguard prevents commands from reading local files in denied
// locations, e.g. the session files when c8y is run by another tool on
// behalf of an AI agent. The denied paths are set from the
// settings.files.deniedPaths setting when the configuration is loaded, e.g. by
// the parent process with the C8Y_SETTINGS_FILES_DENIEDPATHS environment
// variable.
//
// Only reads of files which are referenced by the user (or agent) are guarded,
// e.g. the --data, --template and --file flags and jsonnet imports. Files which
// c8y reads itself, such as its configuration, are not affected.
package fileguard

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// ErrDenied the file is in a denied location
var ErrDenied = errors.New("access to the file is denied")

var (
	mu     sync.RWMutex
	denied []string
)

func parse(paths ...string) []string {
	out := []string{}
	for _, p := range paths {
		if strings.TrimSpace(p) == "" {
			continue
		}
		out = append(out, resolve(p)...)
	}
	return out
}

// resolve returns the absolute path, and the path with symlinks evaluated if it differs
func resolve(p string) []string {
	abs, err := filepath.Abs(p)
	if err != nil {
		abs = filepath.Clean(p)
	}
	out := []string{abs}
	if real, err := evalSymlinks(abs); err == nil && real != abs {
		out = append(out, real)
	}
	return out
}

// evalSymlinks evaluates the symlinks of the path, or of its closest existing parent
// if it does not exist
func evalSymlinks(p string) (string, error) {
	real, err := filepath.EvalSymlinks(p)
	if err == nil {
		return real, nil
	}
	parent := filepath.Dir(p)
	if parent == p {
		return "", err
	}
	realParent, perr := evalSymlinks(parent)
	if perr != nil {
		return "", perr
	}
	return filepath.Join(realParent, filepath.Base(p)), nil
}

// caseInsensitive file systems are case-insensitive by default on macOS and Windows
var caseInsensitive = runtime.GOOS == "darwin" || runtime.GOOS == "windows"

func within(p, dir string) bool {
	if caseInsensitive {
		p, dir = strings.ToLower(p), strings.ToLower(dir)
	}
	if p == dir {
		return true
	}
	if !strings.HasSuffix(dir, string(filepath.Separator)) {
		dir += string(filepath.Separator)
	}
	return strings.HasPrefix(p, dir)
}

// SetDeniedPaths replaces the denied files and folders. Relative paths are
// resolved against the current working directory
func SetDeniedPaths(paths ...string) {
	resolved := parse(paths...)
	mu.Lock()
	defer mu.Unlock()
	denied = resolved
}

func deniedPaths() []string {
	mu.RLock()
	defer mu.RUnlock()
	return denied
}

// Enabled returns true if any paths are denied
func Enabled() bool {
	return len(deniedPaths()) > 0
}

// Check returns an error if the file is in a denied location
func Check(path string) error {
	if path == "" || path == "-" {
		return nil
	}
	d := deniedPaths()
	if len(d) == 0 {
		return nil
	}
	return check(path, d)
}

func check(path string, deniedPaths []string) error {
	for _, p := range resolve(path) {
		for _, d := range deniedPaths {
			if within(p, d) {
				return fmt.Errorf("%w: %s", ErrDenied, path)
			}
		}
	}
	return nil
}

// Open opens the file for reading if it is not in a denied location
func Open(path string) (*os.File, error) {
	if err := Check(path); err != nil {
		return nil, err
	}
	return os.Open(path)
}

// ReadFile reads the file if it is not in a denied location
func ReadFile(path string) ([]byte, error) {
	if err := Check(path); err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}
