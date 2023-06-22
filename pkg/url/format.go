package url

import (
	"net/url"
	"strings"
)

func ParseURL(value string) (*url.URL, error) {
	if !strings.Contains(value, "//") {
		value = "//" + value
	}

	u, err := url.Parse(value)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		switch u.Port() {
		case "443":
			u.Scheme = "https"
			u.Host = u.Hostname()
		case "":
			u.Scheme = "https"
		case "80":
			u.Scheme = "http"
			u.Host = u.Hostname()
		default:
			u.Scheme = "http"
		}
	}

	return u, nil
}
