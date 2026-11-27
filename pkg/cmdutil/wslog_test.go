package cmdutil

import (
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// the texts of the websocket clients in v2.59.0
func TestWebsocketLogHandlerTexts(t *testing.T) {
	tests := []struct {
		msg   string
		attrs []slog.Attr
		want  string
	}{
		{"Establishing connection", []slog.Attr{slog.String("to", "ws://h/cep/realtime")}, "Establishing connection to ws://h/cep/realtime"},
		{"Waiting for message", []slog.Attr{slog.String("id", "1")}, "Waiting for message: id=1"},
		{"ws (send)", []slog.Attr{slog.String("channel", "/meta/connect"), slog.Any("text", []byte(`{"id":"2"}`))}, `ws (send): /meta/connect : {"id":"2"}`},
		{"ws (recv)", []slog.Attr{slog.String("channel", "/meta/connect"), slog.String("text", `{"id":"2"}`)}, `ws (recv): /meta/connect : {"id":"2"}`},
		{"Pending messages", []slog.Attr{slog.Any("id", []string{"4", "3"})}, "Pending messages ids: 4,3"},
		{"Pending messages", []slog.Attr{slog.Any("id", []string{})}, "Pending messages ids: "},
		{"Removing message from pending requests", []slog.Attr{slog.String("id", "7")}, "Removing message from pending requests: 7"},
		{"Received message", []slog.Attr{slog.String("id", "7")}, "Received message 7"},
		{"Resending /meta/connect heartbeat after a small delay", []slog.Attr{slog.Int("delay_ms", 500)}, "Resending /meta/connect heartbeat with 500 ms delay"},
		{"Subscribing to pattern", []slog.Attr{slog.String("value", "/alarms/1")}, "Subscribing to /alarms/1"},
		{"Successfully subscribed to channel", []slog.Attr{slog.String("value", "/alarms/1")}, "Successfully subscribed to channel /alarms/1"},
		{"Token is invalid", []slog.Attr{slog.String("err", "token is malformed")}, "Token is invalid. token is malformed"},
		{"Creating new notification2 token", nil, "Creating new notification2 token"},
		{"Existing token", []slog.Attr{
			slog.String("alg", "HS256"), slog.Bool("valid", false), slog.Bool("expired", false),
			slog.Time("issuedAt", time.Date(2026, 11, 27, 9, 0, 0, 0, time.UTC)), slog.Time("expiresAt", time.Date(2026, 11, 27, 10, 0, 0, 0, time.UTC)),
			slog.String("subscription", "s"), slog.String("subscriber", "r"), slog.Bool("shared", false), slog.String("tenant", "t1"),
		}, "Existing token: alg=HS256, valid=false, expired=false, issuedAt: 2026-11-27 09:00:00 +0000 UTC, expiresAt: 2026-11-27 10:00:00 +0000 UTC, subscription=s, subscriber=r, shared=false, tenant=t1"},
		{"Something else", []slog.Attr{slog.String("a", "1"), slog.String("b", "2")}, "Something else. a: 1, b: 2"},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			attrs := map[string]slog.Value{}
			order := []string{}
			for _, a := range tt.attrs {
				attrs[a.Key] = a.Value
				order = append(order, a.Key)
			}
			assert.Equal(t, tt.want, websocketMessage(tt.msg, attrs, order))
		})
	}
}

func TestWebsocketLogHandlerListensBeforeSubscribingAndCanBeMuted(t *testing.T) {
	log, read := captureLog(t)
	var muted atomic.Bool
	l := slog.New(NewWebsocketLogHandler(log, &muted, "Dropped"))

	l.Info("Subscribing to pattern", "value", "/alarms/1")
	l.Info("Dropped", "a", 1)
	l.Info("Received message", "id", "3")
	muted.Store(true)
	l.Info("Pending messages", "id", []string{"1"})

	out := logPrefix.ReplaceAllString(read(), "")
	assert.Equal(t, "Listening to subscriptions: /alarms/1\nSubscribing to /alarms/1\nReceived message 3\n", out)
}

// the credentials of the realtime handshake are hidden like the ones of the requests
func TestWebsocketLogHandlerHidesTheHandshakeToken(t *testing.T) {
	log, read := captureLog(t)
	h := NewWebsocketLogHandler(log, nil).(*websocketLogHandler)
	h.hideTokens = true
	l := slog.New(h)

	l.Info("ws (send)", "channel", "/meta/handshake", "text", `{"channel":"/meta/handshake","ext":{"com.cumulocity.authn":{"token":"dDEvdTpw"}},"version":"1.0"}`)
	l.Info("ws (send)", "channel", "/meta/handshake", "text", `{"ext":{"com.cumulocity.authn":{"token":"abc","xsrfToken":"def"}}}`)

	out := logPrefix.ReplaceAllString(read(), "")
	assert.Equal(t, `ws (send): /meta/handshake : {"channel":"/meta/handshake","ext":{"com.cumulocity.authn":{"token":"{token}"}},"version":"1.0"}`+"\n"+
		`ws (send): /meta/handshake : {"ext":{"com.cumulocity.authn":{"token":"{token}","xsrfToken":"{token}"}}}`+"\n", out)
}
