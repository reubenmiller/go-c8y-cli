package gjsonpath

import "strings"

// EscapePath escapes a gjson/sjson path
func EscapePath(s string) string {
	// https://github.com/tidwall/sjson/blob/master/sjson.go#L47
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "#", "\\#")
	s = strings.ReplaceAll(s, "@", "\\@")
	s = strings.ReplaceAll(s, "*", "\\*")
	return strings.ReplaceAll(s, "?", "\\?")
}
