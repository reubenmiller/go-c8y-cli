package cmdutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

// captureLog returns a logger which writes info messages to the returned function's result
func captureLog(t *testing.T) (*logger.Logger, func() string) {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stderr
	os.Stderr = w
	log := logger.NewLogger("test", logger.Options{Level: zapcore.InfoLevel})
	os.Stderr = orig
	return log, func() string {
		_ = w.Close()
		out, _ := io.ReadAll(r)
		return string(out)
	}
}

var logPrefix = regexp.MustCompile(`(?m)^\S+\s+\S*INFO\S*\s+`)

func newVerboseTestTransport(t *testing.T, server *httptest.Server, log *logger.Logger) (*verboseTransport, *config.ClientSession) {
	t.Helper()
	f := newClientTestFactory(t, map[string]any{
		"host": server.URL, "tenant": "t123", "username": "peter", "password": "secret-pass",
		config.SettingsLoggerHideSensitive: true,
		config.SettingsViewOption:          "off",
	})
	cfg, err := f.Config()
	require.NoError(t, err)
	baseURL, auth, _ := f.resolveSession(cfg, log)
	f.resolveSession(cfg, log)
	session := config.NewClientSession(baseURL, auth, "")
	return &verboseTransport{base: http.DefaultTransport, cfg: cfg, session: &session, log: log, verbose: true}, &session
}

func TestVerboseTransportLogsLikeV2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"managedObjects":[{"id":"1"},{"id":"2"}],"statistics":{"pageSize":2}}`))
	}))
	defer server.Close()
	log, read := captureLog(t)
	transport, _ := newVerboseTestTransport(t, server, log)

	req, err := http.NewRequest(http.MethodPost, server.URL+"/inventory/managedObjects?pageSize=2", strings.NewReader(`{"name":"x"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("peter", "secret-pass")
	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, `{"managedObjects":[{"id":"1"},{"id":"2"}],"statistics":{"pageSize":2}}`, string(body), "the body is still readable")

	out := logPrefix.ReplaceAllString(read(), "")
	out = regexp.MustCompile(`Response time: \d+ms`).ReplaceAllString(out, "Response time: Nms")
	assert.Equal(t, strings.Join([]string{
		// the session is resolved twice (here and when the transport masks the values), logged once
		"Binding authorization environment variables",
		"Using client auth type: BASIC",
		"Headers: map[Authorization:[Basic  {base64 tenant/username:password}] Content-Type:[application/json]]",
		"Sending request: POST http://{host}/inventory/managedObjects?pageSize=2",
		`Body: {"name":"x"}` + "\n",
		"Status code: 200",
		"Response time: Nms",
		"Response Content-Type: application/json",
		"Response Length: 0.1KB",
		"Unfiltered array size. len=2",
		"View mode: off",
		"",
	}, "\n"), out)
	assert.NotContains(t, out, "secret-pass")
}

func TestVerboseTransportLogsAnErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"inventory/Not Found","message":"No managedObject for id '9'!"}`))
	}))
	defer server.Close()
	log, read := captureLog(t)
	transport, _ := newVerboseTestTransport(t, server, log)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/inventory/managedObjects/9", nil)
	require.NoError(t, err)
	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "No managedObject")

	out := read()
	assert.Contains(t, out, "Invalid response received from server. GET "+server.URL+"/inventory/managedObjects/9: 404 inventory/Not Found No managedObject for id '9'!")
	assert.NotContains(t, out, "Status code:")
}

func TestVerboseTransportMasksABearerToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	log, read := captureLog(t)
	transport, _ := newVerboseTestTransport(t, server, log)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/x", nil)
	require.NoError(t, err)
	// a token which is not the one of the session (e.g. refreshed)
	req.Header.Set("Authorization", "Bearer refreshed-token-value")
	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	out := read()
	assert.NotContains(t, out, "refreshed-token-value")
	assert.Contains(t, out, "Bearer {token}")
}

func TestUnfilteredArraySize(t *testing.T) {
	n, ok := unfilteredArraySize([]byte(`{"managedObjects":[{"id":"1"}],"statistics":{}}`))
	assert.True(t, ok)
	assert.Equal(t, 1, n)

	_, ok = unfilteredArraySize([]byte(`{"id":"1","tags":["a"]}`))
	assert.False(t, ok, "an object")
	_, ok = unfilteredArraySize([]byte(`{"a":[1],"b":[2]}`))
	assert.False(t, ok, "more than one array")
}
