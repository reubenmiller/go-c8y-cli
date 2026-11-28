package cmdutil

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/testclient"

	"github.com/gorilla/websocket"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
)

// newTLSWebsocketServer serves a websocket on /ws with a self-signed certificate
func newTLSWebsocketServer(t *testing.T) *httptest.Server {
	t.Helper()
	upgrader := websocket.Upgrader{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		_ = conn.Close()
	}))
	t.Cleanup(server.Close)
	return server
}

// newWrappedClient creates a client whose transport is wrapped, as the CLI
// wraps it to log the requests. The SDK can not see the TLS settings of the
// transport then, which is why the connectors set them explicitly.
func newWrappedClient(t *testing.T, server *httptest.Server) *apiv2.Client {
	t.Helper()
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return testclient.Must(apiv2.ClientOptions{
		BaseURL:            server.URL,
		Auth:               authentication.AuthOptions{Tenant: "t1", Username: "u", Password: "p"},
		InsecureSkipVerify: true,
		SkipTenantLookup:   true,
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return base.RoundTrip(req)
		}),
	})
}

func dial(t *testing.T, connect func(context.Context, string, http.Header) (*websocket.Conn, *http.Response, error), ctx context.Context, server *httptest.Server) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, _, err := connect(ctx, "wss://"+strings.TrimPrefix(server.URL, "https://")+"/ws", nil)
	if err == nil {
		_ = conn.Close()
	}
	return err
}

func Test_WebsocketConnector_RealtimeSkipsVerificationAlways(t *testing.T) {
	server := newTLSWebsocketServer(t)
	client := newWrappedClient(t, server)

	// the SDK's own connector verifies the certificate: it can not see the setting of the wrapped transport
	err := dial(t, client.DialWebsocket, context.Background(), server)
	if err == nil || !strings.Contains(err.Error(), "certificate") {
		t.Fatalf("expected a certificate error from the default connector, got %v", err)
	}

	// v2.59.0: the realtime websocket never verified the certificate
	connector := NewRealtimeConnector(client, false)
	if err := dial(t, connector.DialWebsocket, context.Background(), server); err != nil {
		t.Fatalf("realtime connector should skip the verification, got %v", err)
	}
}

func Test_WebsocketConnector_ConnectsInDryRun(t *testing.T) {
	server := newTLSWebsocketServer(t)
	client := newWrappedClient(t, server)
	dryCtx := apiv2.WithDryRun(context.Background(), true)

	// v3 before: a dry run did not connect (it returned a dry run error)
	if err := dial(t, client.DialWebsocket, dryCtx, server); err == nil || !strings.Contains(err.Error(), "dry run") {
		t.Fatalf("expected the dry run error from the default connector, got %v", err)
	}

	// v2.59.0: --dry connected
	for name, connector := range map[string]*WebsocketConnector{
		"realtime":      NewRealtimeConnector(client, false),
		"notification2": NewNotification2Connector(client, true),
	} {
		if err := dial(t, connector.DialWebsocket, dryCtx, server); err != nil {
			t.Errorf("%s: should connect in a dry run, got %v", name, err)
		}
	}
}

func Test_WebsocketConnector_Notification2HonorsInsecureSetting(t *testing.T) {
	server := newTLSWebsocketServer(t)
	client := newWrappedClient(t, server)

	if err := dial(t, NewNotification2Connector(client, false).DialWebsocket, context.Background(), server); err == nil || !strings.Contains(err.Error(), "certificate") {
		t.Fatalf("expected a certificate error without the insecure setting, got %v", err)
	}
	if err := dial(t, NewNotification2Connector(client, true).DialWebsocket, context.Background(), server); err != nil {
		t.Fatalf("expected to connect with the insecure setting, got %v", err)
	}
}

func Test_WebsocketConnector_UsesTheClientAuthorization(t *testing.T) {
	var authorization string
	upgrader := websocket.Upgrader{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		if conn, err := upgrader.Upgrade(w, r, nil); err == nil {
			_ = conn.Close()
		}
	}))
	defer server.Close()
	client := newWrappedClient(t, server)

	if err := dial(t, NewRealtimeConnector(client, false).DialWebsocket, context.Background(), server); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(authorization, "Basic ") {
		t.Errorf("expected basic authorization, got %q", authorization)
	}
}

// roundTripperFunc adapts a function to http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func Test_WebsocketConnector_RealtimeHandshakeTimeoutIsTenSeconds(t *testing.T) {
	// v2.59.0: 10 seconds (the default of the SDK is 45)
	connector := NewRealtimeConnector(nil, false)
	if connector.HandshakeTimeout != 10*time.Second {
		t.Fatalf("expected a handshake timeout of 10s, got %v", connector.HandshakeTimeout)
	}
}

// newSilentServer accepts connections and never answers
func newSilentServer(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			t.Cleanup(func() { _ = conn.Close() })
		}
	}()
	t.Cleanup(func() { _ = listener.Close() })
	return "ws://" + listener.Addr().String() + "/cep/realtime"
}

// newBasicClient creates a client for a server which answers everything
func newBasicClient(t *testing.T) *apiv2.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(server.Close)
	return testclient.Must(apiv2.ClientOptions{
		BaseURL:          server.URL,
		Auth:             authentication.AuthOptions{Tenant: "t1", Username: "u", Password: "p"},
		SkipTenantLookup: true,
	})
}

func Test_WebsocketConnector_HandshakeTimeoutApplies(t *testing.T) {
	url := newSilentServer(t)
	client := newBasicClient(t)
	connector := NewRealtimeConnector(client, false)
	connector.HandshakeTimeout = 300 * time.Millisecond

	start := time.Now()
	_, _, err := connector.DialWebsocket(context.Background(), url, nil)
	if err == nil {
		t.Fatal("expected a timeout")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("the handshake timeout did not apply, waited %v", elapsed)
	}
}

func Test_WebsocketConnector_CancelEndsTheHandshake(t *testing.T) {
	url := newSilentServer(t)
	client := newBasicClient(t)
	// the default timeout is much longer than the test
	connector := NewRealtimeConnector(client, false)

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(200*time.Millisecond, cancel)
	start := time.Now()
	_, _, err := connector.DialWebsocket(ctx, url, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected the dial to be cancelled, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("the cancel did not end the handshake, waited %v", elapsed)
	}
}
