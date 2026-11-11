// Package fileguard prevents commands from reading local files in denied
// locations, e.g. the session files when c8y is run by the MCP server on
// behalf of an AI agent. The denied paths are set by the parent process with
// the C8Y_DENIED_PATHS environment variable (separated by os.PathListSeparator),
// so they can't be changed by the command's arguments.
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

// EnvDeniedPaths environment variable which contains the denied paths
const EnvDeniedPaths = "C8Y_DENIED_PATHS"

// ErrDenied the file is in a denied location
var ErrDenied = errors.New("access to the file is denied")

var (
	once   sync.Once
	denied []string
)

func load() {
	denied = parse(os.Getenv(EnvDeniedPaths))
}

func parse(value string) []string {
	out := []string{}
	for _, p := range filepath.SplitList(value) {
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

// SetDeniedPaths replaces the denied paths (instead of reading them from the
// environment). It is used by tests
func SetDeniedPaths(paths ...string) {
	once.Do(func() {})
	denied = parse(strings.Join(paths, string(os.PathListSeparator)))
}

// Enabled returns true if any paths are denied
func Enabled() bool {
	once.Do(load)
	return len(denied) > 0
}

// Check returns an error if the file is in a denied location
func Check(path string) error {
	if !Enabled() || path == "" || path == "-" {
		return nil
	}
	return check(path, denied)
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
