package extsdk

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
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

// Client returns a Cumulocity client for the active c8y session. The credentials are used
// in the following order: token, authorization header, then username without a password
// (which fails on the first request, see CheckError)
func (c *Context) Client() (*c8y.Client, error) {
	if !c.HasSession() {
		return nil, ErrNoSession
	}
	httpClient, err := c.HTTPClient()
	if err != nil {
		return nil, err
	}

	opts := c8y.ClientOptions{
		BaseURL:  c.Host,
		Tenant:   c.Tenant,
		Username: c.Username,
	}
	switch {
	case c.Token != "":
		opts.Token = c.Token
		opts.AuthType = c8y.AuthTypeBearer
	case c.AuthorizationHeader != "":
		// The header is added by the http transport
		opts.AuthType = c8y.AuthTypeNone
	default:
		opts.AuthType = c8y.AuthTypeBasic
	}
	client := c8y.NewClientFromOptions(httpClient, opts)
	client.Version = c.CumulocityVersion
	return client, nil
}

// CheckError adds a hint how to fix authentication errors (401) returned by the Cumulocity client.
// Other errors are returned unchanged
func CheckError(err error) error {
	var errResp *c8y.ErrorResponse
	if errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode() == http.StatusUnauthorized {
		return fmt.Errorf("%w. Authentication failed: %s", err, errExpiredCredentialMsg)
	}
	return err
}
