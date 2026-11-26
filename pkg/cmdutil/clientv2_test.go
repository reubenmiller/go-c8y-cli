package cmdutil

import (
	"net/http"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newClientTestFactory(t *testing.T, values map[string]any) *Factory {
	t.Helper()
	for _, name := range []string{"C8Y_HOST", "C8Y_BASEURL", "C8Y_URL", "C8Y_USER", "C8Y_USERNAME", "C8Y_PASSWORD", "C8Y_TOKEN", "C8Y_TENANT"} {
		t.Setenv(name, "")
	}
	v := viper.New()
	for key, value := range values {
		v.Set(key, value)
	}
	cfg := config.NewConfig(v)
	ios, _, _, _ := iostreams.Test()
	return &Factory{
		IOStreams: ios,
		Config:    func() (*config.Config, error) { return cfg, nil },
		Logger:    func() (*logger.Logger, error) { return cfg.Logger, nil },
	}
}

func TestClientOptionsV2BasicAuth(t *testing.T) {
	f := newClientTestFactory(t, map[string]any{
		"host": "example.com", "tenant": "t123", "username": "peter", "password": "secret",
	})

	opts, err := f.ClientOptionsV2(false)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/", opts.BaseURL)
	assert.Equal(t, "peter", opts.Auth.Username)
	assert.Equal(t, "secret", opts.Auth.Password)
	assert.Equal(t, "t123", opts.Auth.Tenant)
	assert.Equal(t, []authentication.AuthType{authentication.AuthTypeBasic}, opts.Auth.AuthType)
	assert.Empty(t, opts.Auth.Token)
	assert.False(t, opts.InsecureSkipVerify)
	assert.IsType(t, &http.Transport{}, opts.Transport)
}

func TestClientOptionsV2BearerAuth(t *testing.T) {
	f := newClientTestFactory(t, map[string]any{
		"host": "https://example.com", "token": "tok",
		config.SettingsLoginType:     "OAUTH2_INTERNAL",
		"settings.defaults.insecure": true,
	})
	opts, err := f.ClientOptionsV2(true)
	require.NoError(t, err)
	assert.Equal(t, []authentication.AuthType{authentication.AuthTypeBearer}, opts.Auth.AuthType)
	assert.Equal(t, "tok", opts.Auth.Token)
	assert.True(t, opts.InsecureSkipVerify, "the allowInsecure setting is used")
}

func TestClientOptionsV2NoneKeepsTheCredentials(t *testing.T) {
	f := newClientTestFactory(t, map[string]any{
		"host": "https://example.com", "username": "peter", "password": "secret", "tenant": "t123",
		config.SettingsLoginType: "NONE",
	})
	opts, err := f.ClientOptionsV2(false)
	require.NoError(t, err)
	assert.Equal(t, []authentication.AuthType{authentication.AuthTypeNone}, opts.Auth.AuthType)
	// the login can fall back to basic auth
	assert.Equal(t, "peter", opts.Auth.Username)
	assert.Equal(t, "secret", opts.Auth.Password)
	assert.Equal(t, "t123", opts.Auth.Tenant)
}

func TestClientV2NoneSendsNoCredentials(t *testing.T) {
	f := newClientTestFactory(t, map[string]any{
		"host": "https://example.com", "username": "peter", "password": "secret", "tenant": "t123",
		config.SettingsLoginType: "NONE",
	})
	client, err := f.ClientV2()
	require.NoError(t, err)
	assert.Empty(t, client.Auth.Username)
	header, err := client.AuthorizationHeader(nil)
	require.NoError(t, err)
	assert.Empty(t, header)
}

func TestClientV2UsesTheSessionWithoutTheLegacyClient(t *testing.T) {
	f := newClientTestFactory(t, map[string]any{
		"host": "https://example.com", "tenant": "t123", "username": "peter", "password": "secret",
	})
	// Factory.Client is not set, as the v2 client does not depend on it
	client, err := f.ClientV2()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/", client.BaseURL.String())
	assert.Equal(t, "peter", client.Auth.Username)
	header, err := client.AuthorizationHeader(nil)
	require.NoError(t, err)
	assert.Equal(t, "Basic dDEyMy9wZXRlcjpzZWNyZXQ=", header)
}

func TestClientSession(t *testing.T) {
	f := newClientTestFactory(t, map[string]any{
		"host": "example.com:8443", "tenant": "t123", "username": "peter", "password": "secret",
	})
	session, err := f.ClientSession()
	require.NoError(t, err)
	assert.Equal(t, "example.com:8443", session.Host)
	assert.Equal(t, "https://example.com:8443/", session.URL)
	assert.Equal(t, "peter", session.Username)
	assert.Equal(t, authentication.AuthTypeBasic, session.AuthType)
}

func TestClientSessionWithoutHost(t *testing.T) {
	f := newClientTestFactory(t, nil)
	session, err := f.ClientSession()
	require.NoError(t, err)
	assert.Empty(t, session.Host, "a session without a host has not been loaded")
}

// The client sends what v2.59.0 sent: "go-client" as X-Application and
// User-Agent, every request once unless --retries is given, and with
// settings.http.compression=false no Accept-Encoding.
func TestClientOptionsV2RequestSettings(t *testing.T) {
	f := newClientTestFactory(t, map[string]any{
		"host": "example.com", "tenant": "t123", "username": "peter", "password": "secret",
		"settings.http.compression": true,
	})
	opts, err := f.ClientOptionsV2(false)
	require.NoError(t, err)
	assert.Equal(t, "go-client", opts.Agent)
	require.NotNil(t, opts.Retry)
	assert.Equal(t, 0, opts.Retry.Count, "no retries unless --retries is given")
	assert.False(t, opts.Transport.(*http.Transport).DisableCompression)

	f = newClientTestFactory(t, map[string]any{
		"host": "example.com", "tenant": "t123", "username": "peter", "password": "secret",
		config.SettingsHTTPMaxRetries: 2,
		"settings.http.compression":   false,
	})
	opts, err = f.ClientOptionsV2(false)
	require.NoError(t, err)
	assert.Equal(t, 2, opts.Retry.Count)
	assert.True(t, opts.Transport.(*http.Transport).DisableCompression)
}
