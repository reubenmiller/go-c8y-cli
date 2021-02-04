package flags

import (
	"io/ioutil"
	"os"
)

// resolveContents checks whether the given string is a file reference if so it returns the contents, otherwise it returns the
// input value as is
func resolveContents(content string) string {
	if _, err := os.Stat(content); err != nil {
		// not a file
		return content
	}

	fileContent, err := ioutil.ReadFile(content)
	if err != nil {
		return content
	}
	// file contents
	return string(fileContent)
}
