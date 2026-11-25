package cmdutil

import (
	"compress/gzip"
	"crypto/tls"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	c8ycache "github.com/reubenmiller/go-c8y/v2/pkg/c8y"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
)

// sessionFileFlag returns the value of the --session flag
func (f *Factory) sessionFileFlag() string {
	if f.Command == nil {
		return ""
	}
	v, _ := f.Command.PersistentFlags().GetString("session")
	return v
}

// resolveSession resolves the url and the authentication of the session from the configuration
// (and the environment variables if the session is not given using the --session flag).
func (f *Factory) resolveSession(cfg *config.Config, log *logger.Logger) (baseURL *url.URL, auth authentication.AuthOptions, authErrors []error) {
	if f.sessionFileFlag() == "" {
		log.Info("Binding authorization environment variables")
		if err := cfg.BindAuthorization(); err != nil {
			log.Warnf("Failed to bind to authorization variables. %s", err)
		}
	}

	// Note: the session username and password override are already part of the configuration
	clientAuth, authErrors := cfg.LoadClientAuth("", "")
	if clientAuth.Tenant == "" && clientAuth.Token != "" {
		// the tenant is required by some APIs, and tokens contain it
		if claims, err := authentication.ParseToken(clientAuth.Token); err == nil {
			clientAuth.Tenant = claims.Tenant
		}
	}
	baseURL, _ = url.Parse(apiv2.FormatBaseURL(c8ysession.FormatHost(cfg.GetHostWithEnvironment())))
	return baseURL, clientAuth.Auth(), authErrors
}

// ClientSession returns the session values (credentials and host) of the current session.
// The values are loaded from the configuration, and the host environment variables if the session does
// not have a host. It is used to hide the sensitive session values in the log messages.
func (f *Factory) ClientSession() (config.ClientSession, error) {
	cfg, err := f.Config()
	if err != nil {
		return config.ClientSession{}, err
	}
	log, err := f.Logger()
	if err != nil {
		return config.ClientSession{}, err
	}
	baseURL, auth, _ := f.resolveSession(cfg, log)
	return config.NewClientSession(baseURL, auth, cfg.GetCumulocityVersion()), nil
}

// ClientOptionsV2 returns the options to create a go-c8y v2 client for the current session.
// The options contain the credentials and authorization type of the session, and the connection
// settings (proxy, tls verification). It does not include the HTTP cache. The credentials are included
// even if the session does not use them for the requests (the NONE login type).
// A warning is logged if the credentials could not be loaded (e.g. the session is encrypted and
// the passphrase is not known yet), unless the encryption check is disabled (the caller is responsible
// for checking the encryption).
func (f *Factory) ClientOptionsV2(disableEncryptionCheck bool) (apiv2.ClientOptions, error) {
	opts, _, err := f.clientOptionsV2(disableEncryptionCheck)
	return opts, err
}

func (f *Factory) clientOptionsV2(disableEncryptionCheck bool) (apiv2.ClientOptions, config.ClientSession, error) {
	cfg, err := f.Config()
	if err != nil {
		return apiv2.ClientOptions{}, config.ClientSession{}, err
	}
	log, err := f.Logger()
	if err != nil {
		return apiv2.ClientOptions{}, config.ClientSession{}, err
	}

	if cfg.HideSensitive() {
		os.Setenv(apiv2.EnvVarLoggerHideSensitive, "true")
	}

	log.Debug("Creating c8y client")
	cfg.ConfigureProxySettings(log)

	baseURL, auth, authErrors := f.resolveSession(cfg, log)
	if !disableEncryptionCheck && len(authErrors) > 0 {
		log.Warnf("Could not load authentication. error=%v", authErrors[0])
	}
	log.Infof("Using client auth type: %s", auth.GetAuthTypes()[0].String())
	session := config.NewClientSession(baseURL, auth, cfg.GetCumulocityVersion())

	// Note: The authorization which was detected from the session is set (see AuthOptions.AuthType).
	// Without this a password-only session makes the v2 client exchange the password for a
	// bearer token (POST /tenant/oauth/token) on EVERY command, where Authorization: Basic is
	// sent directly. Bearer sessions keep the token-source behaviour so expired tokens still refresh.
	return apiv2.ClientOptions{
		BaseURL:            baseURL.String(),
		Auth:               auth,
		InsecureSkipVerify: cfg.SkipSSLVerify(),
		Transport:          newBaseTransport(cfg),
		Timeout:            cfg.RequestTimeout(),
	}, session, nil
}

// newBaseTransport creates the transport used to connect to Cumulocity
func newBaseTransport(cfg *config.Config) *http.Transport {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: cfg.SkipSSLVerify(),
	}
	if cfg.IgnoreProxy() {
		tr.Proxy = nil
	}
	return tr
}

// ClientV2 builds a go-c8y v2 client from the session. The credentials are resolved from the
// configuration (encrypted session decryption, passphrase prompting, token refresh).
// --verbose/--debug enable the v2 client's request/response logging.
//
// Every call returns a new client, so callers may install their own
// middleware on it (the c8ystream Runner adds activity logging and output
// capture). Shell completion uses it as is.
func (f *Factory) ClientV2() (*apiv2.Client, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}
	opts, session, err := f.clientOptionsV2(false)
	if err != nil {
		return nil, err
	}
	opts.Debug = cfg.Verbose() || cfg.Debug()
	if opts.Auth.GetAuthTypes()[0] == authentication.AuthTypeNone {
		// no credentials are sent
		opts.Auth = authentication.AuthOptions{Tenant: opts.Auth.Tenant, AuthType: opts.Auth.AuthType}
	}

	// Install the HTTP response cache. The cache is a RoundTripper wrapper, so it
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
		// replaying the entry without decompressing would parse
		// gzip bytes as JSON. decompressTransport normalises the response before
		// the cache sees it.
		opts.Transport = c8ycache.NewCachedTransport(
			decompressTransport{base: opts.Transport},
			cfg.CacheTTL(),
			cfg.CacheDir(),
			func(req *http.Request) bool {
				// Never cache authentication requests. The v2 client acquires
				// bearer tokens through this same transport, and
				// a cached token response outlives revocation: the token-source
				// refresh POST is then also served from the cache, so the client
				// re-fetches the same revoked token and every request 401s until
				// the cache TTL expires.
				if strings.Contains(req.URL.Path, "/tenant/oauth") {
					return false
				}
				return strings.Contains(cachableMethods, req.Method)
			},
			c8ycache.CacheOptions{
				ExcludeAuth: !cfg.CacheKeyIncludeAuth(),
				ExcludeHost: !cfg.CacheKeyIncludeHost(),
				Mode:        cfg.CacheMode(),
				BodyKeys:    cfg.CacheBodyKeys(),
			},
		)
	}

	// The SDK logged "Sending request: <method> <url>" (session values masked)
	// before every send. Wrap the transport outside the cache so the line also
	// appears for cache hits; dry runs still skip it because DryRunTransport
	// short-circuits before this layer.
	if f.Logger != nil {
		if logg, logErr := f.Logger(); logErr == nil && logg != nil {
			base := opts.Transport
			opts.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				logg.Infof("Sending request: %s %s", req.Method, cfg.HideSensitiveInformationIfActive(&session, req.URL.String()))
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
	// Parts of the SDK (e.g. the response cache: "Using cached response") log
	// through the default slog logger instead of the client's logger, which
	// would print info messages on stderr of every command.
	slog.SetDefault(logger)

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
