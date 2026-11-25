package cmdutil

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
)

// Helpers for the legacy (go-c8y v1) client. They can be deleted together with Factory.Client.

// LegacyClientSession returns the session values of a legacy client
func LegacyClientSession(client *c8y.Client) *config.ClientSession {
	if client == nil {
		return nil
	}
	session := &config.ClientSession{
		Tenant:     client.TenantName,
		Username:   client.Username,
		Password:   client.Password,
		Token:      client.Token,
		Version:    client.Version,
		OmitTenant: !client.UseTenantInUsername,
	}
	if client.BaseURL != nil {
		session.Host = client.BaseURL.Host
	}
	switch client.AuthorizationType {
	case c8y.AuthTypeNone:
		session.AuthType = authentication.AuthTypeNone
	case c8y.AuthTypeBasic:
		session.AuthType = authentication.AuthTypeBasic
	case c8y.AuthTypeBearer:
		session.AuthType = authentication.AuthTypeBearer
	}
	return session
}

// HideSensitiveForLegacyClient returns a function which hides the sensitive information of the
// session of a legacy client in a message (if the configuration says so)
func HideSensitiveForLegacyClient(cfg *config.Config) func(*c8y.Client, string) string {
	return func(client *c8y.Client, message string) string {
		return cfg.HideSensitiveInformationIfActive(LegacyClientSession(client), message)
	}
}
