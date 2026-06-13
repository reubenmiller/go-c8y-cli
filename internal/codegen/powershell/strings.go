package powershell

import "strings"

// upperFirst capitalizes the first character (invariant culture).
func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// matchesTrueYes reproduces the PowerShell expression ($v -match "true|yes"),
// a case-insensitive substring match.
func matchesTrueYes(v string) bool {
	lower := strings.ToLower(v)
	return strings.Contains(lower, "true") || strings.Contains(lower, "yes")
}
