package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logintype"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
)

// EnvCumulocityHostNames the environment variables which can contain the Cumulocity host,
// as Cumulocity uses different environment variable names for different tooling
var EnvCumulocityHostNames = []string{
	"C8Y_HOST",
	"C8Y_BASEURL",
	"C8Y_URL",
}

// GetHostFromEnvironment gets the first non-empty host environment variable value
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

// GetHostWithEnvironment gets the host of the session. If the session does not have a host,
// then the host environment variables are used.
func (c *Config) GetHostWithEnvironment() string {
	if host := c.GetHost(); host != "" {
		return host
	}
	if host := GetHostFromEnvironment(); host != "" {
		c.Logger.Debugf("Using URL from env variable. %s", host)
		return host
	}
	return ""
}

// ClientAuth the credentials and the authorization type which should be used by a client
type ClientAuth struct {
	Tenant   string
	Username string
	Password string
	Token    string

	// AuthType the authorization used for the requests. It is unset if it can not be detected
	AuthType authentication.AuthType
}

// Auth returns the authentication options of the credentials
func (a ClientAuth) Auth() authentication.AuthOptions {
	opts := authentication.AuthOptions{
		Tenant:   a.Tenant,
		Username: a.Username,
		Password: a.Password,
		Token:    a.Token,
	}
	if a.AuthType != authentication.AuthTypeUnset {
		opts.AuthType = []authentication.AuthType{a.AuthType}
	}
	return opts
}

func splitTenantUsername(username string) (tenant string, user string) {
	if parts := strings.SplitN(username, "/", 2); len(parts) == 2 {
		if parts[0] != "" {
			tenant = parts[0]
		}
		return tenant, parts[1]
	}
	return "", username
}

// LoadClientAuth loads the credentials of the session, and detects the authorization
// which should be used based on the login type. The given username and password
// take precedence over the values of the session.
//
// The returned errors are the problems which prevented the credentials from being
// loaded (e.g. an encrypted password where the passphrase is not available).
func (c *Config) LoadClientAuth(username, password string) (ClientAuth, []error) {
	if username == "" {
		username = c.GetUsername()
	}

	tenant, username := splitTenantUsername(username)
	if tenant == "" {
		tenant = c.GetTenant()
	}

	authErrors := make([]error, 0)
	if password == "" {
		if pass, err := c.GetPassword(); err == nil {
			password = pass
		} else {
			authErrors = append(authErrors, err)
		}
	}

	auth := ClientAuth{
		Tenant:   tenant,
		Username: username,
		Password: password,
	}
	if username != "" && password != "" {
		auth.AuthType = authentication.AuthTypeBasic
	}

	if err := c.loadAuthentication(&auth); err != nil {
		authErrors = append(authErrors, err)
	} else {
		// Clear any existing auth errors
		authErrors = nil
	}
	return auth, authErrors
}

func (a *ClientAuth) setBasic(tenant, username, password string) {
	if t, u, found := strings.Cut(username, "/"); found {
		tenant = t
		username = u
	}
	a.Tenant = tenant
	a.Username = username
	a.Password = password
	a.AuthType = authentication.AuthTypeBasic
}

func (a *ClientAuth) setToken(token string) {
	a.Token = token
	a.AuthType = authentication.AuthTypeBearer
}

func (c *Config) loadAuthentication(auth *ClientAuth) error {
	loginType := c.GetLoginTypeRaw()
	if loginType == "" {
		// auto detect auth method based on the info which is there

		// oauth2_internal / token
		token, err := c.GetToken()
		if err == nil && token != "" {
			auth.setToken(token)
			return nil
		}

		// password
		if p, err := c.GetPassword(); err == nil && p != "" {
			auth.setBasic(c.GetTenant(), c.GetUsername(), p)
			return nil
		}

		// none
		auth.AuthType = authentication.AuthTypeNone
		return nil
	}

	// Force the usage of an auth method regardless if the pre-requisites for
	// such auth method are available (this allows users to also enforce it)
	loginType = c.GetLoginTypeWithDefault()

	// OAUTH2 (internal and external)
	if strings.EqualFold(loginType, logintype.OAuth2Internal) || strings.EqualFold(loginType, logintype.OAuth2) ||
		strings.EqualFold(loginType, logintype.Browser) || strings.EqualFold(loginType, logintype.Certificate) ||
		strings.EqualFold(loginType, logintype.Device) {
		token, err := c.GetToken()
		if err != nil {
			return err
		}
		auth.setToken(token)
		return nil
	}

	if strings.EqualFold(loginType, logintype.Basic) {
		password, err := c.GetPassword()
		if err != nil {
			return err
		}
		auth.setBasic(c.GetTenant(), c.GetUsername(), password)
		return nil
	}

	if strings.EqualFold(loginType, logintype.None) {
		// clear token as we want to force basic auth
		auth.AuthType = authentication.AuthTypeNone
		return nil
	}
	return nil
}

// ConfigureProxySettings configures the proxy environment variables from the configuration.
// Either use explicit proxy, ignore proxy, or use existing env variables
//
//	--proxy "http://10.0.0.1:8080"
//	--noProxy
//	HTTP_PROXY=http://10.0.0.1:8080
//	NO_PROXY=localhost,127.0.0.1
func (c *Config) ConfigureProxySettings(log *logger.Logger) {
	proxy := c.Proxy()
	noProxy := c.IgnoreProxy()
	if noProxy {
		log.Debug("using explicit noProxy setting")
		os.Setenv("HTTP_PROXY", "")
		os.Setenv("HTTPS_PROXY", "")
		os.Setenv("http_proxy", "")
		os.Setenv("https_proxy", "")
		return
	}

	if proxy != "" {
		log.Debugf("using explicit proxy [%s]", proxy)

		os.Setenv("HTTP_PROXY", proxy)
		os.Setenv("HTTPS_PROXY", proxy)
		os.Setenv("http_proxy", proxy)
		os.Setenv("https_proxy", proxy)
		return
	}

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
