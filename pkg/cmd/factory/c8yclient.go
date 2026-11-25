package factory

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/request"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
	"github.com/spf13/viper"
)

// WithCompression sets the compression option for the http client
func WithCompression(enable bool) c8y.ClientOption {
	return func(tr http.RoundTripper) http.RoundTripper {
		tr.(*http.Transport).DisableCompression = !enable
		return tr
	}
}

// RetryLogger to customize the log messages produced by the http retry client
type RetryLogger struct {
	l *logger.Logger
}

func (l RetryLogger) Printf(format string, args ...interface{}) {
	format = strings.TrimPrefix(format, "[DEBUG] ")
	l.l.Infof(format, args...)
}

func CreateCumulocityClient(f *cmdutil.Factory, sessionFile, username, password string, disableEncryptionCheck bool) func() (*c8y.Client, error) {
	return func() (*c8y.Client, error) {
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}
		log, err := f.Logger()
		if err != nil {
			return nil, err
		}
		consol, err := f.Console()
		if err != nil {
			return nil, err
		}

		if cfg.HideSensitive() {
			os.Setenv(c8y.EnvVarLoggerHideSensitive, "true")
		}

		log.Debug("Creating c8y client")
		cfg.ConfigureProxySettings(log)

		internalHttpClient := c8y.NewHTTPClient(
			WithProxyDisabled(cfg.IgnoreProxy()),
			c8y.WithInsecureSkipVerify(cfg.SkipSSLVerify()),
			WithCompression(cfg.ShouldUseCompression()),
		)

		var httpClient *http.Client
		if cfg.HTTPRetryMax() > 0 {
			// Use retry client
			// Note: Don't use the retry client unless we have to as it will the body (if set)
			// of outgoing HTTP requests into memory (and it affects the loading as well)
			retryClient := retryablehttp.NewClient()
			retryClient.RetryMax = cfg.HTTPRetryMax()
			retryClient.RetryWaitMin = cfg.HTTPRetryWaitMin()
			retryClient.RetryWaitMax = cfg.HTTPRetryWaitMax()
			retryClient.Logger = RetryLogger{l: log}
			retryClient.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
				// Pass error back so that the activity log is processed
				log.Warnf("Giving up after %d attempt/s. err=%s", numTries, err)
				if resp != nil && resp.Body != nil {
					defer resp.Body.Close()
				}
				return resp, err
			}
			retryClient.HTTPClient = internalHttpClient
			httpClient = retryClient.StandardClient()
		} else {
			httpClient = internalHttpClient
		}

		cacheBodyPaths := cfg.CacheBodyKeys()
		if len(cacheBodyPaths) > 0 {
			log.Infof("Caching of body only includes paths: %s", strings.Join(cacheBodyPaths, ", "))
		}

		if cfg.CacheEnabled() && cfg.CacheTTL() > 0 {
			cachableMethods := cfg.CacheMethods()
			httpClient = c8y.NewCachedClient(
				httpClient,
				cfg.CacheDir(),
				cfg.CacheTTL(),
				func(r *http.Request) bool {
					return strings.Contains(cachableMethods, r.Method)
				},
				c8y.CacheOptions{
					ExcludeAuth: !cfg.CacheKeyIncludeAuth(),
					ExcludeHost: !cfg.CacheKeyIncludeHost(),
					Mode:        c8y.StoreMode(cfg.CacheMode()),
					BodyKeys:    cacheBodyPaths,
				},
			)
		}

		if sessionFile != "" {
			// Do nothing
		} else {
			log.Info("Binding authorization environment variables")
			if err := cfg.BindAuthorization(); err != nil {
				log.Warnf("Failed to bind to authorization variables. %s", err)
			}
		}

		clientAuth, authErrors := cfg.LoadClientAuth(username, password)

		c8yURL := cfg.GetHostWithEnvironment()

		client := c8y.NewClient(
			httpClient,
			c8ysession.FormatHost(c8yURL),
			clientAuth.Tenant,
			clientAuth.Username,
			clientAuth.Password,
			true,
		)

		if domain := cfg.GetDomain(); domain != "" {
			client.SetDomain(domain)
		}

		if client == nil {
			return nil, fmt.Errorf("failed to create client")
		}

		if client.BaseURL == nil {
			return nil, fmt.Errorf("invalid client url")
		}

		client.SetRequestOptions(c8y.DefaultRequestOptions{
			DryRun: cfg.DryRun(),
			DryRunHandler: func(options *c8y.RequestOptions, req *http.Request) {
				handler := &request.RequestHandler{
					IsTerminal:    f.IOStreams.IsStdoutTTY(),
					IO:            f.IOStreams,
					Client:        client,
					Config:        cfg,
					Logger:        log,
					Console:       consol,
					HideSensitive: cmdutil.HideSensitiveForLegacyClient(cfg),
				}
				handler.DryRunHandler(f.IOStreams, options, req)
			},
		})

		// load authentication
		applyClientAuth(client, clientAuth)

		log.Infof("Using client auth type: %s", client.AuthorizationType.String())

		if !disableEncryptionCheck && len(authErrors) > 0 {
			log.Warnf("Could not load authentication. error=%v", authErrors[0])
		}

		timeout := cfg.RequestTimeout()
		log.Debugf("timeout: %v", timeout)

		// Should we use the tenant in the name or not
		if viper.IsSet("useTenantPrefix") {
			client.UseTenantInUsername = viper.GetBool("useTenantPrefix")
		}

		// Add the realtime client
		client.Realtime = c8y.NewRealtimeClient(
			client.BaseURL.String(),
			newWebsocketDialer(cfg.IgnoreProxy()),
			client.TenantName,
			client.Username,
			client.Password,
		)

		// Set realtime authorization
		if client.AuthorizationType == c8y.AuthTypeBearer {
			if client.Token != "" {
				client.Realtime.SetBearerToken(client.Token)
			} else {
				client.Realtime.SetXSRFToken(client.GetXSRFToken())
			}
		}

		if client.TenantName == "" {
			// Set the tenant either from token, or by looking it up as the tenant is required for a lot of API calls
			log.Debug("Looking up tenant name as it is not set (it is required by some API)")
			client.TenantName = client.GetTenantName(c8y.WithDisabledDryRunContext(context.Background()))

			if client.TenantName == "" {
				log.Info("Failed to lookup tenant name. API calls which require the tenant name will not work!")
			}
		}

		return client, nil
	}
}

// applyClientAuth sets the credentials and the authorization of the session to the legacy client
func applyClientAuth(client *c8y.Client, auth config.ClientAuth) {
	client.TenantName = auth.Tenant
	client.Username = auth.Username
	client.Password = auth.Password

	switch auth.AuthType {
	case authentication.AuthTypeBasic:
		client.SetAuthorizationType(c8y.AuthTypeBasic)
	case authentication.AuthTypeBearer:
		client.SetToken(auth.Token)
	case authentication.AuthTypeNone:
		client.SetAuthorizationType(c8y.AuthTypeNone)
	}
}

func newWebsocketDialer(ignoreProxySettings bool) *websocket.Dialer {
	dialer := &websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  10 * time.Second,
		EnableCompression: false,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	if ignoreProxySettings {
		dialer.Proxy = nil
	}

	return dialer
}

func WithProxyDisabled(disable bool) c8y.ClientOption {
	return func(tr http.RoundTripper) http.RoundTripper {
		if disable {
			tr.(*http.Transport).Proxy = nil
		}
		return tr
	}
}

// clientMissing is the default legacy client function, which is used until the client is configured
func clientMissing() (*c8y.Client, error) {
	return nil, fmt.Errorf("client is missing")
}

// NewLegacyClient returns a function which creates the legacy (go-c8y v1) client once, and returns
// the same client on every call. It is used for the commands which have not been migrated to the
// go-c8y v2 client yet (see cmdutil.Factory.Client).
func NewLegacyClient(f *cmdutil.Factory, sessionFile, username, password string, disableEncryptionCheck bool) func() (*c8y.Client, error) {
	var mu sync.Mutex
	var client *c8y.Client
	return func() (*c8y.Client, error) {
		mu.Lock()
		defer mu.Unlock()

		if client != nil {
			return client, nil
		}
		// the logger must be set before the client is created, as the creation sends requests
		if log, logErr := f.Logger(); logErr == nil && log != nil {
			c8y.Logger = log
		} else {
			c8y.Logger = logger.NewDummyLogger("c8y")
		}
		created, err := CreateCumulocityClient(f, sessionFile, username, password, disableEncryptionCheck)()
		client = created
		return created, err
	}
}
