package url

import "net/url"

// EscapeQuery escapes query parameters so they can be encoded into the URL without conflicting with special characters like "&"
func EscapeQuery(v []byte) []byte {
	q := string(v)
	raw, err := url.QueryUnescape(q)

	if err == nil {
		q = raw
	}

	return []byte(url.QueryEscape(q))
}

// EscapeQuery escapes query parameters so they can be encoded into the URL without conflicting with special characters like "&"
func EscapeQueryString(v string) string {
	raw, err := url.QueryUnescape(v)

	if err == nil {
		v = raw
	}

	return url.QueryEscape(v)
}
