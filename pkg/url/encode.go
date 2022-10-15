package url

import (
	"net/url"
	"strings"
)

// EscapeQuery escapes query parameters so they can be encoded into the URL without conflicting with special characters like "&"
func EscapeQuery(v []byte) []byte {
	return []byte(EscapeQueryString(string(v)))
}

// EscapeQuery escapes query parameters so they can be encoded into the URL without conflicting with special characters like "&"
func EscapeQueryString(v string) string {
	raw, err := url.QueryUnescape(v)

	if err == nil {
		v = raw
	}

	// Preserve special characters
	v = url.QueryEscape(v)
	v = strings.ReplaceAll(v, "%2A", "*")
	v = strings.ReplaceAll(v, "%3A", ":")
	v = strings.ReplaceAll(v, "+", "%2B")
	return v
}
