package c8ystream

import (
	"bytes"
	"io"
	"net/http"

	ctxhelpers "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/contexthelpers"
)

// dryRunTransport guards dry-run requests at the transport layer: the
// currently pinned go-c8y v2 version sends the real request when dry run is
// enabled but mock responses are disabled. Fixed upstream in the v2 dry-run
// middleware — remove this guard once the module is bumped.
type dryRunTransport struct {
	base http.RoundTripper
}

func newDryRunTransport() *dryRunTransport {
	base := http.RoundTripper(http.DefaultTransport)
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		// Clone so client TLS settings never mutate the process-wide default.
		base = t.Clone()
	}
	return &dryRunTransport{base: base}
}

func (t *dryRunTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !ctxhelpers.IsDryRun(req.Context()) {
		return t.base.RoundTrip(req)
	}
	resp := &http.Response{
		Status:     http.StatusText(http.StatusNoContent),
		StatusCode: http.StatusNoContent,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(nil)),
		Request:    req,
	}
	resp.Header.Set("X-Dry-Run", "true")
	return resp, nil
}

// BaseTransport exposes the wrapped transport so the v2 client can apply TLS
// settings through the wrapper.
func (t *dryRunTransport) BaseTransport() http.RoundTripper {
	return t.base
}
