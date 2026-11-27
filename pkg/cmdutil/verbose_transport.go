package cmdutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/tidwall/gjson"
)

// maxLoggedBodySize is the largest request or response body which --verbose reads
// into memory to log it
const maxLoggedBodySize = 10 * 1024 * 1024

// verboseTransport logs the requests and responses in the layout of the v2.59.0
// --verbose output (no HTTP dump):
//
//	Headers: map[...]
//	Sending request: GET http://{host}/inventory/managedObjects?pageSize=1
//	Body: {...}
//	Status code: 200
//	Response time: 3ms
//	Response Content-Type: application/json
//	Response Length: 0.5KB
//	Unfiltered array size. len=1
//	View mode: off
//
// The sensitive session values are masked with config.HideSensitiveInformationIfActive.
type verboseTransport struct {
	base    http.RoundTripper
	cfg     *config.Config
	session *config.ClientSession
	log     *logger.Logger

	// verbose is true when the log messages are shown: only then the bodies are
	// read to log them
	verbose bool
}

func (t *verboseTransport) mask(message string) string {
	return t.cfg.HideSensitiveInformationIfActive(t.session, message)
}

// headers returns the request headers for the log. A bearer token is masked
// even if it is not the one of the session (e.g. it was refreshed after the
// session was loaded); the other session values and basic credentials are masked
// as in v2.59.0.
func (t *verboseTransport) headers(h http.Header) string {
	if t.cfg.HideSensitive() {
		for _, v := range h.Values("Authorization") {
			if strings.HasPrefix(v, "Bearer ") {
				h = h.Clone()
				h.Set("Authorization", "Bearer {token}")
				break
			}
		}
	}
	// v2.59.0 masked the basic credentials with a double space ("Basic  {base64 ...")
	return strings.ReplaceAll(t.mask(fmt.Sprintf("Headers: %v", h)), "Basic {base64", "Basic  {base64")
}

func (t *verboseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.log.Info(t.headers(req.Header))
	t.log.Infof("Sending request: %s %s", req.Method, t.mask(req.URL.String()))
	if t.verbose {
		t.logRequestBody(req)
	}

	start := time.Now()
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		t.log.Infof("ERROR: Request failed. %s", err)
		return resp, err
	}
	duration := time.Since(start)

	var body []byte
	if t.verbose {
		body = t.readResponseBody(resp)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		t.log.Infof("Invalid response received from server. %s", errorSummary(req, resp, body))
	} else {
		t.log.Infof("Status code: %d", resp.StatusCode)
	}
	t.log.Infof("Response time: %dms", duration.Milliseconds())
	t.log.Infof("Response Content-Type: %s", resp.Header.Get("Content-Type"))
	t.log.Debugf("Response Headers: %v", resp.Header)
	if body != nil {
		t.logResponseBody(resp, body)
	}
	return resp, nil
}

// logRequestBody logs the body of the request, without consuming it
func (t *verboseTransport) logRequestBody(req *http.Request) {
	if req.Body == nil || req.Body == http.NoBody {
		return
	}
	if strings.Contains(req.Header.Get("Content-Type"), "multipart/form-data") {
		// Don't print out multi part forms
		return
	}
	var data []byte
	switch {
	case req.GetBody != nil:
		rc, err := req.GetBody()
		if err != nil {
			return
		}
		data, _ = io.ReadAll(io.LimitReader(rc, maxLoggedBodySize))
		_ = rc.Close()
	case req.ContentLength > 0 && req.ContentLength <= maxLoggedBodySize:
		data, _ = io.ReadAll(req.Body)
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(data))
	default:
		// the size is unknown (a stream)
		return
	}
	if strings.Contains(req.Header.Get("Content-Type"), "json") {
		// v2.59.0 logged the JSON as encoded, including the line break which ends it
		t.log.Infof("Body: %s\n", bytes.TrimSpace(data))
		return
	}
	t.log.Infof("Body: %s", bytes.TrimSpace(data))
}

// readResponseBody reads the body of a JSON response, so that it can be logged,
// and puts it back. It returns nil when the body is not read.
func (t *verboseTransport) readResponseBody(resp *http.Response) []byte {
	if resp.Body == nil || resp.Header.Get("Content-Disposition") != "" {
		return nil
	}
	isError := resp.StatusCode < 200 || resp.StatusCode > 299
	if !isError && !strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "json") {
		return nil
	}
	if resp.ContentLength > maxLoggedBodySize {
		return nil
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxLoggedBodySize+1))
	if err != nil || len(data) > maxLoggedBodySize {
		// Hand the unread rest on to the caller
		resp.Body = &readCloser{Reader: io.MultiReader(bytes.NewReader(data), resp.Body), Closer: resp.Body}
		return nil
	}
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(data))
	return data
}

func (t *verboseTransport) logResponseBody(resp *http.Response, body []byte) {
	if len(body) == 0 {
		return
	}
	if resp.ContentLength > -1 {
		t.log.Infof("Response Length: %0.1fKB", float64(resp.ContentLength)/1024)
	} else if resp.Uncompressed {
		t.log.Infof("Response Length: %0.1fKB (uncompressed)", float64(len(body))/1024)
	} else {
		t.log.Infof("Response Length: %0.1fKB", float64(len(body))/1024)
	}
	if !jsonUtilities.IsValidJSON(body) {
		return
	}
	if n, ok := unfilteredArraySize(body); ok {
		t.log.Infof("Unfiltered array size. len=%d", n)
	}
	t.log.Infof("View mode: %s", t.cfg.ViewOption())
}

// unfilteredArraySize returns the number of items of the array which holds the data of a
// response (e.g. the managedObjects), when it can be detected.
func unfilteredArraySize(body []byte) (int, bool) {
	parsed := gjson.ParseBytes(body)
	if parsed.Get("id").Exists() {
		return 0, false
	}
	property := ""
	arrays := 0
	keys := 0
	parsed.ForEach(func(key, value gjson.Result) bool {
		keys++
		if value.IsArray() {
			arrays++
			property = key.String()
		}
		return true
	})
	// more than one array or a lot of properties: not a collection of items
	if arrays != 1 || keys > 10 {
		return 0, false
	}
	return len(parsed.Get(gjson.Escape(property)).Array()), true
}

type readCloser struct {
	io.Reader
	io.Closer
}

// errorSummary is the summary of an error response: "GET <url>: 404 <type> <message>"
func errorSummary(req *http.Request, resp *http.Response, body []byte) string {
	u := *req.URL
	if q := u.Query(); q.Get("client_secret") != "" {
		q.Set("client_secret", "REDACTED")
		u.RawQuery = q.Encode()
	}
	parsed := gjson.ParseBytes(body)
	return fmt.Sprintf("%s %s: %d %s %s", req.Method, &u, resp.StatusCode, parsed.Get("error").String(), parsed.Get("message").String())
}
