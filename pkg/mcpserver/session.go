package mcpserver

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/pathresolver"
)

// SessionKey name used to identify a session, derived from the session file.
// Only known session file extensions are removed, as session names can contain
// dots, e.g. "example.cumulocity.com-firstname.lastname"
func SessionKey(sessionFile string) string {
	if sessionFile == "" {
		return "default"
	}
	name := filepath.Base(sessionFile)
	if ext := filepath.Ext(name); isSessionFileExtension(ext) {
		name = strings.TrimSuffix(name, ext)
	}
	return name
}

func isSessionFileExtension(ext string) bool {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	for _, v := range config.ConfigExtensions {
		if ext == v {
			return true
		}
	}
	return false
}

// ResolveSessionFile returns the absolute path of the session file. A session can be
// given by name (e.g. --session my-tenant), in which case the matching file is looked
// up in the session home directory. The value is returned as is if no file is found.
func ResolveSessionFile(sessionFile string, sessionHome string) string {
	sessionFile = strings.TrimPrefix(sessionFile, "file://")
	if sessionFile == "" {
		return ""
	}
	if info, err := os.Stat(sessionFile); err == nil && !info.IsDir() {
		if abs, err := filepath.Abs(sessionFile); err == nil {
			return abs
		}
		return sessionFile
	}
	if sessionHome != "" && !strings.ContainsRune(sessionFile, os.PathSeparator) {
		matches, _ := pathresolver.ResolvePaths([]string{sessionHome}, sessionFile+".*", config.ConfigExtensions, "ignore")
		for _, match := range matches {
			if isSessionFileExtension(filepath.Ext(match)) && strings.TrimSuffix(filepath.Base(match), filepath.Ext(match)) == sessionFile {
				return match
			}
		}
	}
	return sessionFile
}
