package extensions

import (
	"fmt"
	"regexp"
	"strings"
)

// NamePrefix is the prefix of the extension folders. It's not part of the extension name
const NamePrefix = "c8y-"

// CommandAnnotation is the annotation which marks the root command of an extension. Its value is
// the extension name
const CommandAnnotation = "c8y.extension"

// NameFormat is the format of a normalized extension name
const NameFormat = "^[a-z0-9][a-z0-9-]*$"

var namePattern = regexp.MustCompile(NameFormat)

// NormalizeName returns the extension name without the c8y- prefix and in lowercase.
// An error is returned if the result is not a valid extension name, e.g. it contains
// spaces, underscores, or path separators
func NormalizeName(name string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if strings.HasPrefix(normalized, NamePrefix) {
		normalized = normalized[len(NamePrefix):]
	}
	if !namePattern.MatchString(normalized) {
		return "", fmt.Errorf("invalid extension name %q. The name (without the %q prefix) must only contain letters, numbers, and '-', and start with a letter or number", name, NamePrefix)
	}
	return normalized, nil
}
