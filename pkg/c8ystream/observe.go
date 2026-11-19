package c8ystream

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"time"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
)

// observedResponse is one real (sent) request/response pair as seen by a
// response observer.
type observedResponse struct {
	Response *http.Response
	// Body is the buffered response body for textual (JSON/text) responses. It
	// is nil for binary/streaming bodies, which are passed through unbuffered.
	Body     []byte
	Duration time.Duration
}

// observeResponses returns a client middleware calling fn for every response
// that was actually received from the server. go-c8y installs user middleware
// inside its DryRunTransport, so dry-run and deferred-prepare requests never
// reach it: observers only see real traffic, without checking X-Dry-Run.
//
// Textual bodies are buffered (and replaced with a re-readable copy) so the
// observer can inspect them; binary downloads stream through untouched.
func observeResponses(fn func(observedResponse)) apiv2.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			start := time.Now()
			resp, err := next.RoundTrip(req)
			if err != nil || resp == nil {
				return resp, err
			}
			obs := observedResponse{Response: resp, Duration: time.Since(start)}
			if resp.Body != nil && isTextual(resp.Header.Get("Content-Type")) {
				body, readErr := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if readErr != nil {
					return nil, readErr
				}
				resp.Body = io.NopCloser(bytes.NewReader(body))
				obs.Body = body
			}
			fn(obs)
			return resp, nil
		})
	}
}

func isTextual(contentType string) bool {
	if contentType == "" {
		return true // e.g. 204, or servers omitting the header; bodies are tiny
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	switch {
	case mediaType == "application/json",
		len(mediaType) > 5 && mediaType[len(mediaType)-5:] == "+json",
		len(mediaType) > 5 && mediaType[:5] == "text/":
		return true
	}
	return false
}
