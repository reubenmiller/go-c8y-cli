package config

import (
	"testing"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestParseLoginTypeWithDefault(t *testing.T) {
	for input, expected := range map[string]string{
		"BASIC":           "BASIC",
		"basic":           "BASIC",
		"OAUTH2_INTERNAL": "OAUTH2_INTERNAL",
		"oauth2_internal": "OAUTH2_INTERNAL",
		"OAUTH2":          "OAUTH2",
		"NONE":            "NONE",
		"none":            "NONE",
		// types which are provided by the CLI (not part of the tenant login types)
		"BROWSER":     "BROWSER",
		"device":      "DEVICE",
		"Certificate": "CERTIFICATE",
		// not a login type
		"":        "",
		"unknown": "",
	} {
		assert.Equal(t, expected, ParseLoginTypeWithDefault(input), "input=%q", input)
	}
}

func TestLoginTypeSettings(t *testing.T) {
	c := NewConfig(viper.New())

	assert.Equal(t, "", c.GetLoginTypeRaw())
	assert.Equal(t, "", c.GetLoginTypeWithDefault())

	c.SetLoginType("basic")
	assert.Equal(t, "BASIC", c.GetLoginTypeRaw())

	c.SetLoginType("BROWSER")
	assert.Equal(t, "BROWSER", c.GetLoginTypeWithDefault())

	// invalid values use the default type
	c.SetLoginType("something")
	assert.Equal(t, "OAUTH2_INTERNAL", c.GetLoginTypeWithDefault())
}

func TestLoginTypeIsBasicWhenTheSessionUsernameIsOverridden(t *testing.T) {
	v := viper.New()
	v.Set("settings.defaults.sessionUsername", "user")
	c := NewConfig(v)
	c.SetLoginType("OAUTH2_INTERNAL")
	assert.Equal(t, "BASIC", c.GetLoginTypeRaw())
}

func TestCacheMode(t *testing.T) {
	v := viper.New()
	c := NewConfig(v)
	assert.Equal(t, 0, int(c.CacheMode()), "read and write by default")

	v.Set(SettingsCacheMode, "storeOnly")
	assert.Equal(t, 1, int(c.CacheMode()))
}

func newAuthConfig(values map[string]any) *Config {
	v := viper.New()
	for key, value := range values {
		v.Set(key, value)
	}
	return NewConfig(v)
}

func TestLoadClientAuth(t *testing.T) {
	t.Run("auto detects the token", func(t *testing.T) {
		c := newAuthConfig(map[string]any{"username": "user", "password": "pass", "token": "tok", "tenant": "t1"})
		auth, errs := c.LoadClientAuth("", "")
		assert.Empty(t, errs)
		assert.Equal(t, authentication.AuthTypeBearer, auth.AuthType)
		assert.Equal(t, "tok", auth.Token)
		assert.Equal(t, "user", auth.Username)
		assert.Equal(t, "t1", auth.Tenant)
	})

	t.Run("auto detects the password", func(t *testing.T) {
		c := newAuthConfig(map[string]any{"username": "t2/user", "password": "pass"})
		auth, errs := c.LoadClientAuth("", "")
		assert.Empty(t, errs)
		assert.Equal(t, authentication.AuthTypeBasic, auth.AuthType)
		assert.Equal(t, "user", auth.Username)
		assert.Equal(t, "t2", auth.Tenant)
		assert.Equal(t, "pass", auth.Password)
		assert.Empty(t, auth.Token)
	})

	t.Run("no credentials means no authorization", func(t *testing.T) {
		auth, errs := newAuthConfig(nil).LoadClientAuth("", "")
		assert.Empty(t, errs)
		assert.Equal(t, authentication.AuthTypeNone, auth.AuthType)
	})

	t.Run("basic login type ignores the token", func(t *testing.T) {
		c := newAuthConfig(map[string]any{"username": "user", "password": "pass", "token": "tok", SettingsLoginType: "BASIC"})
		auth, _ := c.LoadClientAuth("", "")
		assert.Equal(t, authentication.AuthTypeBasic, auth.AuthType)
		assert.Equal(t, "pass", auth.Password)
	})

	for _, loginType := range []string{"OAUTH2_INTERNAL", "OAUTH2", "BROWSER", "CERTIFICATE", "DEVICE"} {
		t.Run(loginType+" login type uses the token", func(t *testing.T) {
			c := newAuthConfig(map[string]any{"username": "user", "password": "pass", "token": "tok", SettingsLoginType: loginType})
			auth, _ := c.LoadClientAuth("", "")
			assert.Equal(t, authentication.AuthTypeBearer, auth.AuthType)
			assert.Equal(t, "tok", auth.Token)
		})
	}

	t.Run("NONE login type sends no authorization", func(t *testing.T) {
		c := newAuthConfig(map[string]any{"username": "user", "password": "pass", SettingsLoginType: "NONE"})
		auth, _ := c.LoadClientAuth("", "")
		assert.Equal(t, authentication.AuthTypeNone, auth.AuthType)
	})

	t.Run("given username and password are used for the initial credentials", func(t *testing.T) {
		// With the NONE login type, the credentials are not replaced by the ones of the session
		c := newAuthConfig(map[string]any{"username": "user", "password": "pass", "tenant": "t1", SettingsLoginType: "NONE"})
		auth, _ := c.LoadClientAuth("t9/other", "secret")
		assert.Equal(t, "other", auth.Username)
		assert.Equal(t, "t9", auth.Tenant)
		assert.Equal(t, "secret", auth.Password)
	})

	t.Run("auth options", func(t *testing.T) {
		auth := ClientAuth{Tenant: "t1", Username: "u", Password: "p", AuthType: authentication.AuthTypeBasic}
		opts := auth.Auth()
		assert.Equal(t, []authentication.AuthType{authentication.AuthTypeBasic}, opts.AuthType)
		assert.Equal(t, "u", opts.Username)

		assert.Empty(t, ClientAuth{Username: "u"}.Auth().AuthType, "an unknown authorization is not set")
	})
}

func TestHostWithEnvironment(t *testing.T) {
	t.Setenv("C8Y_HOST", "")
	t.Setenv("C8Y_BASEURL", "")
	t.Setenv("C8Y_URL", "https://example.com")
	assert.Equal(t, "https://example.com", newAuthConfig(nil).GetHostWithEnvironment())
	assert.Equal(t, "https://session.com", newAuthConfig(map[string]any{"host": "https://session.com"}).GetHostWithEnvironment())
}

func TestClientSessionAuthorizationHeader(t *testing.T) {
	for name, tc := range map[string]struct {
		session  ClientSession
		expected string
	}{
		"bearer":           {ClientSession{Token: "tok", AuthType: authentication.AuthTypeBearer}, "Bearer tok"},
		"bearer no token":  {ClientSession{Username: "u", Password: "p", AuthType: authentication.AuthTypeBearer}, ""},
		"basic":            {ClientSession{Tenant: "t1", Username: "u", Password: "p", AuthType: authentication.AuthTypeBasic}, "Basic dDEvdTpw"},
		"basic no tenant":  {ClientSession{Username: "u", Password: "p", AuthType: authentication.AuthTypeBasic}, "Basic dTpw"},
		"basic omit":       {ClientSession{Tenant: "t1", Username: "u", Password: "p", AuthType: authentication.AuthTypeBasic, OmitTenant: true}, "Basic dTpw"},
		"basic no pass":    {ClientSession{Username: "u", AuthType: authentication.AuthTypeBasic}, ""},
		"none":             {ClientSession{Username: "u", Password: "p", Token: "tok", AuthType: authentication.AuthTypeNone}, ""},
		"unset token":      {ClientSession{Username: "u", Password: "p", Token: "tok"}, "Bearer tok"},
		"unset password":   {ClientSession{Username: "u", Password: "p"}, "Basic dTpw"},
		"unset no secrets": {ClientSession{}, ""},
	} {
		assert.Equal(t, tc.expected, tc.session.AuthorizationHeader(), name)
	}
}

func TestHideSensitiveInformation(t *testing.T) {
	c := NewConfig(viper.New())
	session := &ClientSession{Host: "example.com:8080", Tenant: "t123", Username: "peter", Password: "secret", Token: "tok.en.value"}

	message := "https://example.com:8080/tenant t123/peter:secret Basic dDEyMy9wZXRlcjpzZWNyZXQ= tok.en.value"
	hidden := c.HideSensitiveInformation(session, message)
	assert.NotContains(t, hidden, "example.com")
	assert.NotContains(t, hidden, "t123")
	assert.NotContains(t, hidden, "peter")
	assert.NotContains(t, hidden, "secret")
	assert.NotContains(t, hidden, "dDEyMy9w")
	assert.NotContains(t, hidden, "tok.en.value")
	assert.Contains(t, hidden, "{host}")

	assert.Equal(t, message, c.HideSensitiveInformation(nil, message), "nothing is hidden without a session")
}
