package c8ylogin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logintype"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testToken = "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJhZG1pbiIsInRlbiI6InQxMjM0NSJ9.sig"

// fakeTenant is a minimal Cumulocity server which supports the login options, tokens (with optional two-factor
// authentication), the current tenant and the version.
type fakeTenant struct {
	t *httptest.Server

	mu            sync.Mutex
	loginTypes    []string
	user          string
	password      string
	requireTOTP   string // the required tfa_code, when set
	tokenRequests int
	tfaCodes      []string
	authHeaders   []string
}

func newFakeTenant(t *testing.T, loginTypes ...string) *fakeTenant {
	t.Helper()
	f := &fakeTenant{loginTypes: loginTypes, user: "admin", password: "secret"}
	if len(f.loginTypes) == 0 {
		f.loginTypes = []string{"OAUTH2_INTERNAL", "BASIC"}
	}
	f.t = httptest.NewServer(http.HandlerFunc(f.handle))
	t.Cleanup(f.t.Close)
	return f
}

func (f *fakeTenant) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   "general/internalError",
		"message": message,
		"info":    "https://cumulocity.com/api/core/",
	})
}

func (f *fakeTenant) handle(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch {
	case r.URL.Path == "/tenant/loginOptions":
		options := make([]map[string]any, 0)
		for _, loginType := range f.loginTypes {
			options = append(options, map[string]any{
				"type":        loginType,
				"initRequest": "/tenant/oauth?tenant_id=t12345",
				"self":        "http://t12345.example.com/tenant/loginOptions/" + loginType,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"loginOptions": options})

	case r.URL.Path == "/tenant/oauth/token":
		f.tokenRequests++
		_ = r.ParseForm()
		code := r.Form.Get("tfa_code")
		f.tfaCodes = append(f.tfaCodes, code)
		if r.Form.Get("username") != f.user || r.Form.Get("password") != f.password {
			f.writeError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		validCode := code == f.requireTOTP
		if f.requireTOTP == "*" {
			// any code which looks like a TOTP code
			validCode = len(code) == 6 && strings.Trim(code, "0123456789") == ""
		}
		if f.requireTOTP != "" && !validCode {
			f.writeError(w, http.StatusUnauthorized, "TFA TOTP code required")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": testToken})

	case r.URL.Path == "/tenant/currentTenant":
		auth := r.Header.Get("Authorization")
		f.authHeaders = append(f.authHeaders, auth)
		valid := auth == "Bearer "+testToken
		if user, pass, ok := r.BasicAuth(); ok && strings.HasSuffix(user, f.user) && pass == f.password {
			valid = true
		}
		if !valid {
			f.writeError(w, http.StatusUnauthorized, "Bad credentials")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"name": "t12345", "domainName": "t12345.example.com"})

	case strings.HasPrefix(r.URL.Path, "/tenant/system/options/system/version"):
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"category": "system", "key": "version", "value": "1018.0.0"})

	case r.URL.Path == "/user/currentUser":
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"userName": f.user})

	default:
		http.NotFound(w, r)
	}
}

func newTestHandler(t *testing.T, server *fakeTenant, auth authentication.AuthOptions) *LoginHandlerV2 {
	t.Helper()
	ios, _, _, _ := iostreams.Test()
	saved := 0
	h := NewLoginHandlerV2(ios, apiv2.ClientOptions{
		BaseURL: server.t.URL,
		Auth:    auth,
	}, ios.ErrOut, func() { saved++ })
	return h
}

func TestLoginOAuth2Internal(t *testing.T) {
	server := newFakeTenant(t)
	h := newTestHandler(t, server, authentication.AuthOptions{
		Username: "admin",
		Password: "secret",
		AuthType: []authentication.AuthType{authentication.AuthTypeBearer},
	})

	require.NoError(t, h.Run())

	session := h.Session()
	assert.Equal(t, testToken, session.Token)
	assert.Equal(t, "admin", session.Username)
	assert.Equal(t, "t12345", session.Tenant)
	assert.Equal(t, "1018.0.0", session.Version)
	assert.Equal(t, authentication.AuthTypeBearer, session.AuthType)
	assert.Equal(t, logintype.OAuth2Internal, h.LoginType)
	assert.Equal(t, "Bearer "+testToken, session.AuthorizationHeader())
	assert.Equal(t, 1, server.tokenRequests)
}

func TestLoginBasicUsesTheLoginOptionOfTheTenant(t *testing.T) {
	server := newFakeTenant(t, "BASIC")
	h := newTestHandler(t, server, authentication.AuthOptions{
		Username: "admin",
		Password: "secret",
		AuthType: []authentication.AuthType{authentication.AuthTypeBasic},
	})
	h.LoginType = logintype.Basic

	require.NoError(t, h.Run())

	session := h.Session()
	assert.Empty(t, session.Token)
	assert.Equal(t, "t12345", session.Tenant)
	assert.Equal(t, "1018.0.0", session.Version)
	assert.Equal(t, authentication.AuthTypeBasic, session.AuthType)
	assert.Equal(t, logintype.Basic, h.LoginType)
	assert.Zero(t, server.tokenRequests, "basic auth must not request a token")
	assert.True(t, strings.HasPrefix(session.AuthorizationHeader(), "Basic "))
}

func TestLoginAllowedLoginTypesLimitTheLoginOptions(t *testing.T) {
	server := newFakeTenant(t, "OAUTH2_INTERNAL", "BASIC")
	h := newTestHandler(t, server, authentication.AuthOptions{
		Username: "admin",
		Password: "secret",
	})
	h.LoginType = logintype.Basic
	h.AllowedLoginTypes = []string{"basic"}

	require.NoError(t, h.Run())
	assert.Empty(t, h.Session().Token)
	assert.Zero(t, server.tokenRequests)
}

func TestLoginExistingTokenIsVerifiedAndTheLoginTypeDetected(t *testing.T) {
	server := newFakeTenant(t)
	h := newTestHandler(t, server, authentication.AuthOptions{
		Username: "admin",
		Token:    testToken,
		AuthType: []authentication.AuthType{authentication.AuthTypeBearer},
	})

	require.NoError(t, h.Run())

	assert.Zero(t, server.tokenRequests, "an existing token must be reused")
	assert.Equal(t, testToken, h.Session().Token)
	// the subject of the token is the username, so it was issued by Cumulocity
	assert.Equal(t, logintype.OAuth2Internal, h.LoginType)
	assert.Equal(t, "t12345", h.Session().Tenant)
}

func TestLoginUsernameIsLoadedWhenMissing(t *testing.T) {
	server := newFakeTenant(t)
	h := newTestHandler(t, server, authentication.AuthOptions{
		Token:    testToken,
		AuthType: []authentication.AuthType{authentication.AuthTypeBearer},
	})

	require.NoError(t, h.Run())
	assert.Equal(t, "admin", h.Session().Username)
}

func TestLoginTOTPCodeIsUsedForTheTwoFactorChallenge(t *testing.T) {
	server := newFakeTenant(t)
	server.requireTOTP = "123456"
	h := newTestHandler(t, server, authentication.AuthOptions{
		Username: "admin",
		Password: "secret",
		AuthType: []authentication.AuthType{authentication.AuthTypeBearer},
	})
	h.TOTPCode = "123456"

	require.NoError(t, h.Run())

	assert.Equal(t, testToken, h.Session().Token)
	assert.Equal(t, []string{"", "123456"}, server.tfaCodes, "the code is only sent after the server asks for it")
}

func TestLoginTOTPSecretGeneratesTheCode(t *testing.T) {
	server := newFakeTenant(t)
	h := newTestHandler(t, server, authentication.AuthOptions{
		Username: "admin",
		Password: "secret",
		AuthType: []authentication.AuthType{authentication.AuthTypeBearer},
	})
	h.TOTPSecret = "JBSWY3DPEHPK3PXP"
	// the code can't be predicted, so the server accepts any 6 digit code
	server.requireTOTP = "*"

	require.NoError(t, h.Run())
	assert.Equal(t, testToken, h.Session().Token)
	require.Len(t, server.tfaCodes, 2)
	assert.Len(t, server.tfaCodes[1], 6)
}

func TestLoginTwoFactorCodeRequiresATerminal(t *testing.T) {
	server := newFakeTenant(t)
	server.requireTOTP = "123456"
	h := newTestHandler(t, server, authentication.AuthOptions{
		Username: "admin",
		Password: "secret",
		AuthType: []authentication.AuthType{authentication.AuthTypeBearer},
	})

	// no code was given and the tests do not run in a terminal
	err := h.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "two-factor authentication (TOTP) code is required")
	assert.Empty(t, h.Session().Token)
}

func TestLoginBadCredentialsWithoutPromptDoesNotReturnAToken(t *testing.T) {
	server := newFakeTenant(t)
	h := newTestHandler(t, server, authentication.AuthOptions{
		Username: "admin",
		Password: "wrong",
		AuthType: []authentication.AuthType{authentication.AuthTypeBearer},
	})

	_ = h.Run()
	assert.Empty(t, h.Session().Token)
}

func TestLoginNoLoginOptionsFallsBackToBasic(t *testing.T) {
	server := newFakeTenant(t)
	server.loginTypes = []string{}
	h := newTestHandler(t, server, authentication.AuthOptions{
		Username: "admin",
		Password: "secret",
	})
	h.LoginType = logintype.OAuth2Internal

	require.NoError(t, h.Run())
	assert.Equal(t, logintype.Basic, h.LoginType)
	assert.Equal(t, "t12345", h.Session().Tenant)
}

func TestPreferenceFromLoginType(t *testing.T) {
	assert.Nil(t, PreferenceFromLoginType(""))
	assert.Equal(t, "OAUTH2_INTERNAL", string(PreferenceFromLoginType(logintype.OAuth2Internal)[0]))
	assert.Equal(t, "BASIC", string(PreferenceFromLoginType(logintype.None)[0]))
	assert.Equal(t, "OAUTH2_DEVICE_FLOW", string(PreferenceFromLoginType(logintype.OAuth2)[0]))
	assert.Equal(t, "OAUTH2_DEVICE_FLOW", string(PreferenceFromLoginType(logintype.Device)[0]))
	assert.Equal(t, "OAUTH2_BROWSER_FLOW", string(PreferenceFromLoginType(logintype.Browser)[0]))
	assert.Equal(t, "CERTIFICATE", string(PreferenceFromLoginType(logintype.Certificate)[0]))
}

func TestIsSDKFlow(t *testing.T) {
	server := newFakeTenant(t)
	h := newTestHandler(t, server, authentication.AuthOptions{})
	for loginType, expected := range map[string]bool{
		"":                        false,
		logintype.Basic:           false,
		logintype.OAuth2Internal:  false,
		logintype.OAuth2:          false,
		logintype.None:            false,
		logintype.Browser:         true,
		logintype.Device:          true,
		logintype.Certificate:     true,
		strings.ToLower("DEVICE"): true,
	} {
		h.LoginType = loginType
		assert.Equal(t, expected, h.isSDKFlow(), loginType)
	}
}
