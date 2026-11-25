// Package rasession opens the remote access sessions of the remoteaccess
// server and connect commands: it selects the device's passthrough
// configuration and creates the go-c8y remote access client for it.
package rasession

import (
	"context"
	"strings"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	racfg "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/remoteaccess/remoteaccess_configurations"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/remoteaccess"
)

// Options selects the configuration of a session.
type Options struct {
	// DeviceID is the id of the device (already resolved)
	DeviceID string
	// Configuration is a configuration id, name or name pattern with *
	// wildcards. Empty selects one automatically (see racfg.Service.Detect).
	Configuration string
	// Multiplex, when set, overrides the mux option of the configuration name
	// (e.g. "https+mux:router")
	Multiplex *bool
}

// Session is a remote access client for a device's configuration.
type Session struct {
	*remoteaccess.RemoteAccessClient
	DeviceID      string
	Configuration jsonmodels.RemoteAccessConfiguration
	Multiplex     bool
}

// Detect selects the device's passthrough configuration. Call it with a
// resolution context so it also runs in a dry run.
func Detect(ctx context.Context, client *apiv2.Client, opts Options) (jsonmodels.RemoteAccessConfiguration, error) {
	return client.RemoteAccess.Configurations.Detect(ctx, racfg.DetectOptions{
		ManagedObjectID: opts.DeviceID,
		Name:            opts.Configuration,
	})
}

// New creates the remote access client for a configuration from Detect.
func New(client *apiv2.Client, opts Options, config jsonmodels.RemoteAccessConfiguration) *Session {
	multiplex := false
	if opts.Multiplex != nil {
		multiplex = *opts.Multiplex
	} else if name, ok := remoteaccess.ParseConfigurationName(config.Name()); ok {
		multiplex = name.Multiplex
	}
	return &Session{
		RemoteAccessClient: remoteaccess.NewRemoteAccessClient(client, remoteaccess.RemoteAccessOptions{
			ManagedObjectID: opts.DeviceID,
			RemoteAccessID:  config.ID(),
			Multiplex:       multiplex,
		}),
		DeviceID:      opts.DeviceID,
		Configuration: config,
		Multiplex:     multiplex,
	}
}

// HostPort splits the listener address into host and port. An address
// without a host (e.g. ":2222") uses 127.0.0.1.
func (s *Session) HostPort() (host, port string) {
	host, port, _ = strings.Cut(s.GetListenerAddress(), ":")
	if host == "" {
		host = "127.0.0.1"
	}
	return host, port
}

// BrowserScheme is the URL scheme to open the local proxy with: the scheme
// given, or with "auto" the scheme in the configuration name (e.g. "https" for
// "https:router"), else http. This is the naming convention of
// https://github.com/Cumulocity-IoT/cumulocity-remote-access-cloud-http-proxy
func (s *Session) BrowserScheme(scheme string) string {
	if scheme != "auto" {
		return scheme
	}
	if name, ok := remoteaccess.ParseConfigurationName(s.Configuration.Name()); ok {
		return name.Scheme
	}
	return "http"
}

// CumulocityURL is the tenant URL, for messages
func CumulocityURL(client *apiv2.Client) string {
	return strings.TrimRight(client.BaseURL.String(), "/")
}
