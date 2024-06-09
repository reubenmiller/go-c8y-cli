package artifact

import (
	"path/filepath"
	"regexp"
)

// Parse the name from a given file by striping any version information from the file name
func ParseName(file string) string {
	baseFileName := filepath.Base(file)
	fileExt := filepath.Ext(baseFileName)
	baseFileName = baseFileName[0 : len(baseFileName)-len(fileExt)]

	// Strip suffixes which are added by OS's when downloading file which already exists in the
	// target directory.
	// e.g. "./cloud-http-proxy (1).zip"
	suffixRegex := regexp.MustCompile(`\s+\(\d+\)$`)
	baseFileName = suffixRegex.ReplaceAllString(baseFileName, "")

	versionRegex := regexp.MustCompile(`([_-]v?\d+\.\d+\.\d+(-SNAPSHOT)?)?$`)
	return versionRegex.ReplaceAllString(baseFileName, "")
}
