package config

import (
	"encoding/base64"
	"net/url"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
)

// ClientSession are the session values of a Cumulocity client: the values which are
// hidden in log messages, written to the session file and exported as environment variables.
type ClientSession struct {
	// URL is the base url of Cumulocity including the scheme
	URL string

	// Host is the host (and port) of the Cumulocity url, without the scheme
	Host     string
	Tenant   string
	Username string
	Password string
	Token    string
	Version  string

	// AuthType is the authorization which is used for the requests
	AuthType authentication.AuthType

	// OmitTenant leaves out the tenant from the username when using basic authorization
	OmitTenant bool
}

// NewClientSession creates the session values from the base url and the authentication options of a client
func NewClientSession(baseURL *url.URL, auth authentication.AuthOptions, version string) ClientSession {
	session := ClientSession{
		Tenant:   auth.Tenant,
		Username: auth.Username,
		Password: auth.Password,
		Token:    auth.Token,
		Version:  version,
		AuthType: auth.GetAuthTypes()[0],
	}
	if baseURL != nil {
		session.URL = baseURL.String()
		session.Host = baseURL.Host
	}
	return session
}

// AuthorizationHeader returns the value of the Authorization header which is used for
// the requests, or an empty value if no authorization is used.
func (s ClientSession) AuthorizationHeader() string {
	authType := s.AuthType
	if authType == authentication.AuthTypeUnset {
		if s.Token != "" {
			authType = authentication.AuthTypeBearer
		} else if s.Username != "" && s.Password != "" {
			authType = authentication.AuthTypeBasic
		}
	}

	switch authType {
	case authentication.AuthTypeBearer:
		if s.Token != "" {
			return "Bearer " + s.Token
		}
	case authentication.AuthTypeBasic:
		username := s.Username
		if !s.OmitTenant {
			username = authentication.JoinTenantUser(s.Tenant, s.Username)
		}
		if username != "" && s.Password != "" {
			return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+s.Password))
		}
	}
	return ""
}
