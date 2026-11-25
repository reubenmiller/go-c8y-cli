package cmdutil

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
)

// ClientV2 builds a go-c8y v2 client from the session via the legacy factory
// client, which owns credential resolution (encrypted session decryption,
// passphrase prompting, token refresh), so both clients always agree.
// --verbose/--debug enable the v2 client's request/response logging.
//
// Every call returns a new client, so callers may install their own
// middleware on it (the c8ystream Runner adds activity logging and output
// capture). Shell completion uses it as is.
func (f *Factory) ClientV2() (*apiv2.Client, error) {
	legacy, err := f.Client()
	if err != nil {
		return nil, err
	}
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}
	opts := apiv2.ClientOptions{
		BaseURL: legacy.BaseURL.String(),
		Auth: authentication.AuthOptions{
			Tenant:   legacy.TenantName,
			Username: legacy.Username,
			Password: legacy.Password,
			Token:    legacy.Token,
		},
		Debug:   cfg.Verbose() || cfg.Debug(),
		Timeout: cfg.RequestTimeout(),
	}

	// Mirror the legacy client's negotiated authorization method. Without this
	// a password-only session makes the v2 client exchange the password for a
	// bearer token (POST /tenant/oauth/token) on EVERY command, where v1 sent
	// Authorization: Basic directly — slower, chattier, and a poisoning hazard
	// when the exchange is replayed (revoked tokens). Bearer sessions keep the
	// token-source behaviour so expired tokens still refresh.
	switch legacy.AuthorizationType {
	case c8y.AuthTypeBasic:
		opts.Auth.AuthType = []authentication.AuthType{authentication.AuthTypeBasic}
	case c8y.AuthTypeBearer:
		opts.Auth.AuthType = []authentication.AuthType{authentication.AuthTypeBearer}
	}

	// Install the HTTP response cache exactly as the legacy factory client does
	// (pkg/cmd/factory/c8yclient.go). The cache is a RoundTripper wrapper, so it
	// becomes the v2 client's base transport; apiv2.NewClient further wraps it in
	// DryRunTransport, giving DryRun -> cache -> network ordering (a dry run
	// short-circuits before ever touching the cache). Without this the v2 path
	// silently skipped caching entirely.
	if cfg.CacheEnabled() && cfg.CacheTTL() > 0 {
		cachableMethods := cfg.CacheMethods()
		// The cache must store DECODED bodies. The v2 resty client sends an
		// explicit Accept-Encoding header, which disables Go's transparent gzip
		// handling, so responses reach this transport still compressed with
		// Content-Encoding: gzip. Cached as-is they poison the shared cache dir:
		// the v1-path commands (whose transport decompresses transparently before
		// their cache wrapper) replay the entry without decompressing and parse
		// gzip bytes as JSON. decompressTransport normalises the response before
		// the cache sees it.
		opts.Transport = c8y.CacheResponse(
			cfg.CacheTTL(),
			cfg.CacheDir(),
			func(req *http.Request) bool {
				// Never cache authentication requests. The v2 client acquires
				// bearer tokens through this same transport (v1 never did), and
				// a cached token response outlives revocation: the token-source
				// refresh POST is then also served from the cache, so the client
				// re-fetches the same revoked token and every request 401s until
				// the cache TTL expires.
				if strings.Contains(req.URL.Path, "/tenant/oauth") {
					return false
				}
				return strings.Contains(cachableMethods, req.Method)
			},
			c8y.CacheOptions{
				ExcludeAuth: !cfg.CacheKeyIncludeAuth(),
				ExcludeHost: !cfg.CacheKeyIncludeHost(),
				Mode:        cfg.CacheMode(),
				BodyKeys:    cfg.CacheBodyKeys(),
			},
		)(decompressTransport{base: http.DefaultTransport})
	}

	// v1's SDK logged "Sending request: <method> <url>" (session values masked)
	// before every send. Wrap the transport outside the cache so the line also
	// appears for cache hits, exactly as v1's pre-transport log did; dry runs
	// still skip it because DryRunTransport short-circuits before this layer.
	if f.Logger != nil {
		if logg, logErr := f.Logger(); logErr == nil && logg != nil {
			base := opts.Transport
			if base == nil {
				base = http.DefaultTransport
			}
			opts.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				logg.Infof("Sending request: %s %s", req.Method, cfg.HideSensitiveInformationIfActive(legacy, req.URL.String()))
				return base.RoundTrip(req)
			})
		}
	}

	// The SDK's HTTP layer logs its warnings (e.g. credentials over plain
	// HTTP) at info level and the realtime client logs every websocket frame
	// at info level: show them only with --verbose/--debug.
	level := slog.LevelWarn
	if cfg.Verbose() || cfg.Debug() {
		level = slog.LevelDebug
	}
	var errOut io.Writer = io.Discard
	if f.IOStreams != nil && f.IOStreams.ErrOut != nil {
		errOut = f.IOStreams.ErrOut
	}
	logger := slog.New(slog.NewTextHandler(errOut, &slog.HandlerOptions{Level: level}))
	opts.Logger = logger

	client := apiv2.NewClient(opts)
	client.Realtime().SetLogger(logger.With("component", "realtime"))
	return client, nil
}

// roundTripperFunc adapts a function to http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// decompressTransport gunzips gzip-encoded response bodies (removing the
// Content-Encoding header) so layers above it — most importantly the response
// cache, whose entries are shared with the v1-path commands — always see the
// decoded body, matching what Go's transport produces when no explicit
// Accept-Encoding header suppresses its transparent gzip handling.
type decompressTransport struct {
	base http.RoundTripper
}

func (t decompressTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil || resp == nil {
		return resp, err
	}
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		gz, gzErr := gzip.NewReader(resp.Body)
		if gzErr != nil {
			// Leave the response untouched (e.g. an empty body); the caller's
			// own decoding will surface any real corruption.
			return resp, nil
		}
		resp.Body = &gunzipReadCloser{gz: gz, raw: resp.Body}
		resp.Header.Del("Content-Encoding")
		resp.Header.Del("Content-Length")
		resp.ContentLength = -1
		resp.Uncompressed = true
	}
	return resp, nil
}

// gunzipReadCloser reads decompressed content and closes both the gzip reader
// and the underlying network body.
type gunzipReadCloser struct {
	gz  *gzip.Reader
	raw io.ReadCloser
}

func (g *gunzipReadCloser) Read(p []byte) (int, error) { return g.gz.Read(p) }

func (g *gunzipReadCloser) Close() error {
	err := g.gz.Close()
	if rawErr := g.raw.Close(); err == nil {
		err = rawErr
	}
	return err
}
