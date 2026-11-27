package cmdutil

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/wsdial"
)

// WebsocketConnector opens the websockets of the commands which stream from
// Cumulocity (the realtime commands and notification2 subscriptions) like
// v2.59.0 did, which differs from the default of the SDK's client in two ways:
//
//   - A dry run (--dry) connects: v2.59.0 streamed for the whole --duration
//     (the requests which create a notification2 token are still only shown).
//     The SDK's client does not connect in a dry run and returns at once.
//   - The TLS and proxy settings are given here. The SDK's client takes them
//     from the transport of its HTTP client, which it can not look into when
//     the CLI wraps that transport (e.g. to log the requests): the dialer then
//     verifies every certificate and always uses the proxy of the environment.
//
// The authorization, the request headers and the middleware of the client are
// used as for any request.
type WebsocketConnector struct {
	Client *apiv2.Client

	// InsecureSkipVerify skips the verification of the certificate of the server
	InsecureSkipVerify bool

	// IgnoreProxy connects directly, without the proxy of the environment
	IgnoreProxy bool
}

var _ wsdial.Connector = (*WebsocketConnector)(nil)

// WebsocketHeader returns the request headers of the connection (see
// api.Client.WebsocketHeader).
func (c *WebsocketConnector) WebsocketHeader(ctx context.Context) (http.Header, error) {
	return c.Client.WebsocketHeader(apiv2.WithDryRun(ctx, false))
}

// DialWebsocket connects, also when ctx is in a dry run.
func (c *WebsocketConnector) DialWebsocket(ctx context.Context, rawURL string, header http.Header) (*websocket.Conn, *http.Response, error) {
	ctx = apiv2.WithDryRun(ctx, false)
	if !c.InsecureSkipVerify && !c.IgnoreProxy {
		return c.Client.DialWebsocket(ctx, rawURL, header)
	}

	dialURL, err := websocketURL(rawURL)
	if err != nil {
		return nil, nil, err
	}
	requestHeader, err := c.Client.WebsocketHeader(ctx)
	if err != nil {
		return nil, nil, err
	}
	// the header of the caller wins, e.g. an Authorization header
	for key, values := range header {
		requestHeader.Del(key)
		for _, value := range values {
			requestHeader.Add(key, value)
		}
	}

	dialer := c.Client.WebsocketDialer()
	if c.InsecureSkipVerify {
		if dialer.TLSClientConfig == nil {
			dialer.TLSClientConfig = &tls.Config{}
		}
		dialer.TLSClientConfig.InsecureSkipVerify = true
	}
	if c.IgnoreProxy {
		dialer.Proxy = nil
	}
	return dialer.DialContext(ctx, dialURL, requestHeader)
}

// websocketURL turns an http(s) URL into a ws(s) URL
func websocketURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid websocket url: %w", err)
	}
	switch strings.ToLower(u.Scheme) {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
	default:
		return "", fmt.Errorf("invalid websocket url %q: unsupported scheme %q", rawURL, u.Scheme)
	}
	return u.String(), nil
}

// NewRealtimeConnector returns the connector of the realtime client, which
// like in v2.59.0 never verifies the certificate of the server (whatever the
// insecure setting) and does not use the proxy with --ignoreProxy.
func NewRealtimeConnector(client *apiv2.Client, ignoreProxy bool) *WebsocketConnector {
	return &WebsocketConnector{Client: client, InsecureSkipVerify: true, IgnoreProxy: ignoreProxy}
}

// NewNotification2Connector returns the connector of the notification2
// clients, which like in v2.59.0 verify the certificate of the server unless
// the insecure setting is on.
func NewNotification2Connector(client *apiv2.Client, insecure bool) *WebsocketConnector {
	return &WebsocketConnector{Client: client, InsecureSkipVerify: insecure}
}
