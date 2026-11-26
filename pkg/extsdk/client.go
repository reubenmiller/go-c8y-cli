package extsdk

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
)

// ErrNoSession is returned when a client is requested, but c8y does not have an active session
var ErrNoSession = errors.New("no active c8y session. Activate one using 'set-session' or 'c8y sessions login'")

// headerTransport adds a fixed Authorization header to each request
type headerTransport struct {
	value string
	base  http.RoundTripper
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", t.value)
	return t.base.RoundTrip(clone)
}

// HTTPClient returns the http client used by Client. TLS verification follows the c8y
// insecure setting, and a client certificate is used if the session has one
func (c *Context) HTTPClient() (*http.Client, error) {
	defaultTransport := http.DefaultTransport.(*http.Transport)
	tr := defaultTransport.Clone()
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: c.Insecure,
	}
	if c.Certificate != "" && c.CertificateKey != "" {
		cert, err := tls.LoadX509KeyPair(c.Certificate, c.CertificateKey)
		if err != nil {
			return nil, fmt.Errorf("could not load the session's client certificate: %w", err)
		}
		tr.TLSClientConfig.Certificates = []tls.Certificate{cert}
	}

	var transport http.RoundTripper = tr
	if c.Token == "" && c.AuthorizationHeader != "" {
		transport = &headerTransport{value: c.AuthorizationHeader, base: tr}
	}
	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Minute,
	}, nil
}

// Client returns a Cumulocity client (go-c8y v2) for the active c8y session. The credentials are
// used in the following order: token, authorization header, then username without a password
// (which fails on the first request, see CheckError)
func (c *Context) Client() (*api.Client, error) {
	if !c.HasSession() {
		return nil, ErrNoSession
	}
	httpClient, err := c.HTTPClient()
	if err != nil {
		return nil, err
	}

	auth := authentication.AuthOptions{
		Tenant:   c.Tenant,
		Username: c.Username,
	}
	switch {
	case c.Token != "":
		auth.Token = c.Token
		auth.AuthType = []authentication.AuthType{authentication.AuthTypeBearer}
	case c.AuthorizationHeader != "":
		// The header is added by the http transport
		auth.AuthType = []authentication.AuthType{authentication.AuthTypeNone}
	default:
		auth.AuthType = []authentication.AuthType{authentication.AuthTypeBasic}
	}
	return api.NewClient(api.ClientOptions{
		BaseURL:   c.Host,
		Auth:      auth,
		Transport: httpClient.Transport,
		Timeout:   httpClient.Timeout,
		// the tenant of the session is used, and it must not send any request when creating the client
		SkipTenantLookup: true,
	}), nil
}

// CheckError adds a hint how to fix authentication errors (401) returned by the Cumulocity client.
// Other errors are returned unchanged
func CheckError(err error) error {
	if core.ErrHasStatus(err, http.StatusUnauthorized) {
		return fmt.Errorf("%w. Authentication failed: %s", err, errExpiredCredentialMsg)
	}
	return err
}
