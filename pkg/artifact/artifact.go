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
	versionRegex := regexp.MustCompile(`([_-]v?\d+\.\d+\.\d+(-SNAPSHOT)?)?$`)
	return versionRegex.ReplaceAllString(baseFileName, "")
}
