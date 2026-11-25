package c8ysession

import (
	"bytes"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSession() *CumulocitySession {
	return &CumulocitySession{
		Host:      "https://t123.example.com",
		Tenant:    "t123",
		Username:  "peter",
		Password:  "secret",
		Token:     "token-value",
		LoginType: "OAUTH2_INTERNAL",
		Mode:      "dev",
		Path:      "/tmp/session.json",
	}
}

func bearerClient() *config.ClientSession {
	return &config.ClientSession{
		Tenant:   "t123",
		Username: "peter",
		Password: "secret",
		Token:    "token-value",
		Version:  "1018.0.0",
		AuthType: authentication.AuthTypeBearer,
	}
}

func basicClient() *config.ClientSession {
	return &config.ClientSession{
		Tenant:   "t123",
		Username: "peter",
		Password: "secret",
		Version:  "1018.0.0",
		AuthType: authentication.AuthTypeBasic,
	}
}

func TestGetVariablesFromSessionBearer(t *testing.T) {
	cfg := config.NewConfig(viper.New())
	vars := GetVariablesFromSession(newTestSession(), cfg, bearerClient(), false)

	assert.Equal(t, "https://t123.example.com", vars["C8Y_HOST"])
	assert.Equal(t, "https://t123.example.com", vars["C8Y_URL"])
	assert.Equal(t, "https://t123.example.com", vars["C8Y_BASEURL"])
	assert.Equal(t, "t123.example.com", vars["C8Y_DOMAIN"])
	assert.Equal(t, "t123", vars["C8Y_TENANT"])
	assert.Equal(t, "1018.0.0", vars["C8Y_VERSION"])
	assert.Equal(t, "peter", vars["C8Y_USERNAME"])
	assert.Equal(t, "peter", vars["C8Y_USER"])
	assert.Equal(t, "token-value", vars["C8Y_TOKEN"])
	assert.Equal(t, "OAUTH2_INTERNAL", vars["C8Y_SETTINGS_LOGIN_TYPE"])
	assert.Equal(t, "Bearer token-value", vars["C8Y_HEADER_AUTHORIZATION"])
	assert.Equal(t, "Authorization: Bearer token-value", vars["C8Y_HEADER"])
	assert.Equal(t, "/tmp/session.json", vars["C8Y_SESSION"])
	assert.Equal(t, "dev", vars[config.EnvSessionMode])
	assert.Equal(t, "", vars["C8Y_PASSWORD"], "the password is not needed when a token is used")

	// the password is only included when it is requested
	vars = GetVariablesFromSession(newTestSession(), cfg, bearerClient(), true)
	assert.Equal(t, "secret", vars["C8Y_PASSWORD"])
}

func TestGetVariablesFromSessionBasic(t *testing.T) {
	cfg := config.NewConfig(viper.New())
	session := newTestSession()
	session.Token = ""
	session.LoginType = "BASIC"
	vars := GetVariablesFromSession(session, cfg, basicClient(), false)

	// base64(t123/peter:secret)
	assert.Equal(t, "Basic dDEyMy9wZXRlcjpzZWNyZXQ=", vars["C8Y_HEADER_AUTHORIZATION"])
	assert.Equal(t, "Authorization: Basic dDEyMy9wZXRlcjpzZWNyZXQ=", vars["C8Y_HEADER"])
	assert.Equal(t, "secret", vars["C8Y_PASSWORD"])
	assert.Equal(t, "", vars["C8Y_TOKEN"])
}

func TestGetVariablesFromSessionDoesNotExportTokenForBasicAuth(t *testing.T) {
	cfg := config.NewConfig(viper.New())
	// the session still has a token, but the requests use basic authorization
	vars := GetVariablesFromSession(newTestSession(), cfg, basicClient(), false)
	assert.Equal(t, "", vars["C8Y_TOKEN"])
	assert.True(t, strings.HasPrefix(vars["C8Y_HEADER_AUTHORIZATION"].(string), "Basic "))
}

func TestGetVariablesFromSessionNoAuth(t *testing.T) {
	cfg := config.NewConfig(viper.New())
	client := &config.ClientSession{AuthType: authentication.AuthTypeNone, Username: "peter", Password: "secret"}
	vars := GetVariablesFromSession(newTestSession(), cfg, client, true)
	assert.Equal(t, "", vars["C8Y_HEADER_AUTHORIZATION"])
	assert.Equal(t, "Authorization: ", vars["C8Y_HEADER"])
}

func TestWriteOutput(t *testing.T) {
	cfg := config.NewConfig(viper.New())

	t.Run("bash", func(t *testing.T) {
		out := &bytes.Buffer{}
		require.NoError(t, WriteOutput(out, bearerClient(), cfg, newTestSession(), "bash"))
		assert.Contains(t, out.String(), "export C8Y_HOST='https://t123.example.com'")
		assert.Contains(t, out.String(), "export C8Y_TOKEN='token-value'")
		assert.Contains(t, out.String(), "export C8Y_HEADER_AUTHORIZATION='Bearer token-value'")
		assert.Contains(t, out.String(), "unset C8Y_PASSWORD")
	})

	t.Run("dotenv skips empty values", func(t *testing.T) {
		out := &bytes.Buffer{}
		require.NoError(t, WriteOutput(out, bearerClient(), cfg, newTestSession(), "dotenv"))
		assert.Contains(t, out.String(), "C8Y_TENANT=t123\n")
		assert.Contains(t, out.String(), "C8Y_VERSION=1018.0.0\n")
		assert.NotContains(t, out.String(), "C8Y_PASSWORD=")
		assert.NotContains(t, out.String(), "C8Y_CERTIFICATE=")
	})

	t.Run("json does not include the password when there is a token", func(t *testing.T) {
		out := &bytes.Buffer{}
		require.NoError(t, WriteOutput(out, bearerClient(), cfg, newTestSession(), "json"))
		assert.Contains(t, out.String(), `"token":"token-value"`)
		assert.NotContains(t, out.String(), `"password"`)
	})

	t.Run("no output format", func(t *testing.T) {
		out := &bytes.Buffer{}
		require.NoError(t, WriteOutput(out, bearerClient(), cfg, newTestSession(), ""))
		assert.Empty(t, out.String())
	})

	t.Run("unsupported format", func(t *testing.T) {
		assert.Error(t, WriteOutput(&bytes.Buffer{}, bearerClient(), cfg, newTestSession(), "xml"))
	})
}

func TestPrintSessionInfo(t *testing.T) {
	cfg := config.NewConfig(viper.New())

	t.Run("shows the authorization type and the login type", func(t *testing.T) {
		out := &bytes.Buffer{}
		PrintSessionInfo(out, bearerClient(), cfg, *newTestSession())
		assert.Contains(t, out.String(), "authType     : BEARER (loginType=OAUTH2_INTERNAL)")
		assert.Contains(t, out.String(), "host         : https://t123.example.com")
		assert.Contains(t, out.String(), "username     : peter")
		assert.Contains(t, out.String(), "Cumulocity Session ")
	})

	t.Run("basic", func(t *testing.T) {
		out := &bytes.Buffer{}
		session := newTestSession()
		session.LoginType = ""
		PrintSessionInfo(out, basicClient(), cfg, *session)
		assert.Contains(t, out.String(), "authType     : BASIC\n")
	})

	t.Run("without a client", func(t *testing.T) {
		out := &bytes.Buffer{}
		PrintSessionInfo(out, nil, cfg, *newTestSession())
		assert.NotContains(t, out.String(), "authType")
		assert.Contains(t, out.String(), "tenant       : t123")
	})

	t.Run("sensitive information can be hidden", func(t *testing.T) {
		hidden := config.NewConfig(viper.New())
		hidden.Set(config.SettingsSessionHide, true)
		out := &bytes.Buffer{}
		client := bearerClient()
		client.Host = "t123.example.com"
		PrintSessionInfo(out, client, hidden, *newTestSession())
		assert.Contains(t, out.String(), "sensitive info is hidden")
		assert.Contains(t, out.String(), "tenant       : {tenant}")
		assert.Contains(t, out.String(), "username     : {username}")
		assert.NotContains(t, out.String(), "peter")
	})
}

func TestLoginTypeRequiresToken(t *testing.T) {
	for loginType, expected := range map[string]bool{
		"OAUTH2_INTERNAL": true,
		"oauth2":          true,
		"BROWSER":         true,
		"DEVICE":          true,
		"CERTIFICATE":     true,
		"BASIC":           false,
		"NONE":            false,
		"":                false,
	} {
		assert.Equal(t, expected, LoginTypeRequiresToken(loginType), loginType)
	}
}

func TestGetTokenSubject(t *testing.T) {
	// {"sub":"admin","ten":"t12345"}
	subject, err := GetTokenSubject("eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJhZG1pbiIsInRlbiI6InQxMjM0NSJ9.sig")
	require.NoError(t, err)
	assert.Equal(t, "admin", subject)

	_, err = GetTokenSubject("not-a-token")
	assert.Error(t, err)
}
