package fileutilities

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Matches control characters (0x00-0x1F, 0x7F) AND Windows forbidden chars: < > : " / \ | ? *
var badPathChars = regexp.MustCompile(`[\x00-\x1F\x7F<>:"/\\|?*]`)

func SanitizeFileName(userInput string) string {
	// Standardize path formatting (removes redundant slashes or '.' markers)
	cleaned := filepath.Clean(userInput)

	// Wipe out Windows-illegal characters (except path separators) and control characters
	sanitized := badPathChars.ReplaceAllString(cleaned, "_")

	return sanitized
}

// WriteToFile write input to a filepath which can include template variables
// @text	text to be written to file
// @filename	filename
// @directory	output directory. If empty, then a temp directory will be used
// if filename
func WriteToFile(src io.Reader, filename string, append bool, newline bool, fields map[string]string) (string, error) {
	// Support simple variable substitution to be able to set the output file name dynamically to download a collection of files
	if strings.Contains(filename, "{") && strings.Contains(filename, "}") {
		for k, v := range fields {
			fieldName := fmt.Sprintf("{%s}", k)
			if strings.Contains(filename, fieldName) {
				filename = strings.ReplaceAll(filename, fieldName, SanitizeFileName(v))
			}
		}
	}

	var out *os.File
	var err error
	dirPath := path.Dir(filename)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("could not create directory. dir=%s,  err=%w", dirPath, err)
	}
	if append {
		out, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		out, err = os.Create(filename)
	}

	if err != nil {
		return "", fmt.Errorf("Could not create file. %s", err)
	}
	defer out.Close()

	if append && newline {
		if fs, err := out.Stat(); err == nil {
			if fs.Size() > 0 {
				// add newline when appending so that content is separated (only if file is not empty)
				fmt.Fprintf(out, "\n")
			}
		}
	}

	// Write to file
	_, writeErr := io.Copy(out, src)
	if writeErr != nil {
		return "", fmt.Errorf("failed to copy file contents to file. %s", err)
	}

	if fullpath, err := filepath.Abs(filename); err == nil {
		return fullpath, nil
	}
	return filename, nil
}
