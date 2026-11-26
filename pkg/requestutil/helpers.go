package requestutil

import (
	"net/http"
	"strings"
)

// methodsWithBody are the HTTP methods whose requests can have a body
var methodsWithBody = []string{
	http.MethodDelete,
	http.MethodPatch,
	http.MethodPost,
	http.MethodPut,
}

// RequestSupportsBody checks if the method supports a body with the request
func RequestSupportsBody(method string) bool {
	for _, v := range methodsWithBody {
		if strings.EqualFold(method, v) {
			return true
		}
	}
	return false
}

// FlattenArrayMap converts multi value maps (e.g. query parameters and headers) to a map where
// the values which only have one item are not an array
func FlattenArrayMap[K string, V []string](m map[K]V) map[K]any {
	out := make(map[K]any)
	for key, value := range m {
		if len(value) == 1 {
			out[key] = value[0]
		} else {
			out[key] = value
		}
	}
	return out
}

// HasJSONHeader returns true if the header contains a json content type
func HasJSONHeader(h *http.Header) bool {
	if h == nil {
		return true
	}
	contentType := h.Get("Content-Type")
	if contentType == "" {
		return true
	}
	return strings.Contains(strings.ToLower(contentType), "json")
}
