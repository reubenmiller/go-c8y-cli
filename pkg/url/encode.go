package url

import (
	"fmt"
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

func PathEscape(v string) string {
	raw, err := url.PathUnescape(v)
	if err == nil {
		return url.PathEscape(raw)
	}
	return url.PathEscape(v)
}

func PathEscapePreserveQueryParameters(v string) string {
	var value string
	raw, err := url.PathUnescape(v)
	if err == nil {
		value = raw
	} else {
		value = v
	}

	// strip query parameters
	pathValue := url.PathEscape(value)
	if p1, p2, found := strings.Cut(value, "?"); found {
		pathValue = fmt.Sprintf("%s?%s", url.PathEscape(p1), p2)
	}
	// preserve special characters
	pathValue = strings.ReplaceAll(pathValue, "%2F", "/")
	return pathValue
}
