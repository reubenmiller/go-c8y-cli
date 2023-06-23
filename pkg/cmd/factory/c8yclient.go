package factory

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/request"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/viper"
)

var EnvCumulocityHostNames = []string{
	"C8Y_HOST",
	"C8Y_BASEURL",
	"C8Y_URL",
}

// GetHostFromEnvironment gets the first non-empty host environment variable value
// as Cumulocity uses different environment variable names for different tooling
func GetHostFromEnvironment() string {
	var value = ""
	for _, name := range EnvCumulocityHostNames {
		value = strings.TrimSpace(os.Getenv(name))
		if value != "" {
			break
		}
	}
	return value
}

// WithCompression sets the compression option for the http client
func WithCompression(enable bool) c8y.ClientOption {
	return func(tr http.RoundTripper) http.RoundTripper {
		tr.(*http.Transport).DisableCompression = !enable
		return tr
	}
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
		configureProxySettings(cfg, log)

		httpClient := c8y.NewHTTPClient(
			WithProxyDisabled(cfg.IgnoreProxy()),
			c8y.WithInsecureSkipVerify(cfg.SkipSSLVerify()),
			WithCompression(false),
		)

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
					Mode:        cfg.CacheMode(),
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

		if username == "" {
			username = cfg.GetUsername()
		}

		tenant := ""
		if parts := strings.SplitN(username, "/", 2); len(parts) == 2 {
			if parts[0] != "" {
				tenant = parts[0]
			}
			username = parts[1]
		}
		if tenant == "" {
			tenant = cfg.GetTenant()
		}
		if password == "" {
			pass, err := cfg.GetPassword()
			if !disableEncryptionCheck && err != nil {
				return nil, err
			}
			password = pass
		}

		c8yURL := cfg.GetHost()
		c8yURLFromEnv := GetHostFromEnvironment()
		if c8yURL == "" && c8yURLFromEnv != "" {
			// Get url from env variable if it is empty
			log.Debugf("Using URL from env variable. %s", c8yURLFromEnv)
			c8yURL = c8yURLFromEnv
		}

		client := c8y.NewClient(
			httpClient,
			c8ysession.FormatHost(c8yURL),
			tenant,
			username,
			password,
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
					HideSensitive: cfg.HideSensitiveInformationIfActive,
				}
				handler.DryRunHandler(f.IOStreams, options, req)
			},
		})

		// load authentication
		if err := loadAuthentication(cfg, client); !disableEncryptionCheck && err != nil {
			log.Warnf("Could not load authentication. %s", err)
			return nil, err
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
		if client.AuthorizationMethod == c8y.AuthMethodOAuth2Internal {
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

func loadAuthentication(v *config.Config, c *c8y.Client) error {
	token, err := v.GetToken()
	if err != nil {
		return err
	}
	if token != "" {
		c.SetToken(token)
		c.AuthorizationMethod = c8y.AuthMethodOAuth2Internal
	}
	return nil
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

func configureProxySettings(cfg *config.Config, log *logger.Logger) {

	// Proxy settings
	// Either use explicit proxy, ignore proxy, or use existing env variables
	// --proxy "http://10.0.0.1:8080"
	// --noProxy
	// HTTP_PROXY=http://10.0.0.1:8080
	// NO_PROXY=localhost,127.0.0.1
	proxy := cfg.Proxy()
	noProxy := cfg.IgnoreProxy()
	if noProxy {
		log.Debug("using explicit noProxy setting")
		os.Setenv("HTTP_PROXY", "")
		os.Setenv("HTTPS_PROXY", "")
		os.Setenv("http_proxy", "")
		os.Setenv("https_proxy", "")
	} else {
		if proxy != "" {
			log.Debugf("using explicit proxy [%s]", proxy)

			os.Setenv("HTTP_PROXY", proxy)
			os.Setenv("HTTPS_PROXY", proxy)
			os.Setenv("http_proxy", proxy)
			os.Setenv("https_proxy", proxy)

		} else {
			proxyVars := []string{"HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy", "NO_PROXY", "no_proxy"}

			var proxySettings strings.Builder

			for _, name := range proxyVars {
				if v := os.Getenv(name); v != "" {
					proxySettings.WriteString(fmt.Sprintf(" %s [%s]", name, v))
				}
			}
			if proxySettings.Len() > 0 {
				log.Debugf("Using existing env variables.%s", proxySettings.String())
			}
		}
	}
}
