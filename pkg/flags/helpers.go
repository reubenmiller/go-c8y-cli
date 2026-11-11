package flags

import (
	"os"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileguard"
)

// resolveContents checks whether the given string is a file reference if so it returns the contents, otherwise it returns the
// input value as is. An error is returned if the file is in a denied location
func resolveContents(content string) (string, error) {
	if _, err := os.Stat(content); err != nil {
		// not a file
		return content, nil
	}

	fileContent, err := fileguard.ReadFile(content)
	if err != nil {
		if fileguard.Check(content) != nil {
			return "", err
		}
		return content, nil
	}
	// file contents
	return string(fileContent), nil
}

func isFile(p string) bool {
	if fp, err := os.Stat(p); err == nil && !fp.IsDir() && strings.ContainsAny(p, `/\`) {
		return true
	}
	return false
}
