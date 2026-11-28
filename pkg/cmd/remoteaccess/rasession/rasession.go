// Package rasession opens the remote access sessions of the remoteaccess
// server and connect commands: it selects the device's passthrough
// configuration and creates the go-c8y remote access client for it.
package rasession

import (
	"context"
	"errors"
	"fmt"
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
	*remoteaccess.Client
	DeviceID      string
	Configuration jsonmodels.RemoteAccessConfiguration
	Multiplex     bool
}

// Detect selects the device's passthrough configuration. Call it with a
// resolution context so it also runs in a dry run.
//
// The errors for a missing configuration have the text of v2.59.0
// ("configuration not found. name=<name>"), which the commands print.
func Detect(ctx context.Context, client *apiv2.Client, opts Options) (jsonmodels.RemoteAccessConfiguration, error) {
	config, err := client.RemoteAccess.Configurations.Detect(ctx, racfg.DetectOptions{
		ManagedObjectID: opts.DeviceID,
		Name:            opts.Configuration,
	})
	if errors.Is(err, racfg.ErrPassthroughNotFound) || errors.Is(err, racfg.ErrNoPassthrough) {
		return config, notFoundError(opts.Configuration)
	}
	return config, err
}

// notFoundError is the error of v2.59.0 for a configuration which was not
// found: it is the same when a name was given and when the device has no
// passthrough configuration at all (then the name is empty).
func notFoundError(name string) error {
	return fmt.Errorf("configuration not found. name=%s", name)
}

// New creates the remote access client for a configuration from Detect.
func New(client *apiv2.Client, opts Options, config jsonmodels.RemoteAccessConfiguration) (*Session, error) {
	multiplex := false
	if opts.Multiplex != nil {
		multiplex = *opts.Multiplex
	} else if name, ok := remoteaccess.ParseConfigurationName(config.Name()); ok {
		multiplex = name.Multiplex
	}
	raClient, err := remoteaccess.NewClient(client, remoteaccess.ClientOptions{
		ManagedObjectID: opts.DeviceID,
		RemoteAccessID:  config.ID(),
		Multiplex:       multiplex,
	})
	if err != nil {
		return nil, err
	}
	return &Session{
		Client:        raClient,
		DeviceID:      opts.DeviceID,
		Configuration: config,
		Multiplex:     multiplex,
	}, nil
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
	return strings.TrimRight(client.BaseURL().String(), "/")
}
