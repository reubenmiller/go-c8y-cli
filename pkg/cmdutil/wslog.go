package cmdutil

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
)

// websocketLogHandler is the slog handler of the websocket clients of the SDK
// (realtime and notification2). It writes the records to the CLI logger with the
// texts of v2.59.0 ("ws (send): /meta/connect : {...}", "Pending messages ids: 1,2"),
// not as slog records (time=... level=INFO msg=... component=realtime).
//
// Records which have no text of v2.59.0 are written as "message. key: value".
type websocketLogHandler struct {
	log   *logger.Logger
	attrs []slog.Attr

	// muted stops the log (when it is set), and drop lists the messages which are not logged
	muted *atomic.Bool
	drop  map[string]bool
}

// NewWebsocketLogHandler returns the slog handler which logs the websocket
// clients of the SDK to the CLI logger in the format of v2.59.0. The log stops
// when muted is set, and the records with a message of drop are not logged.
func NewWebsocketLogHandler(log *logger.Logger, muted *atomic.Bool, drop ...string) slog.Handler {
	h := &websocketLogHandler{log: log, muted: muted, drop: map[string]bool{}}
	for _, msg := range drop {
		h.drop[msg] = true
	}
	return h
}

// NewRealtimeLogHandler is the handler of the realtime client of the factory
// (see Factory.MuteWebsocketLog).
func (f *Factory) NewRealtimeLogHandler(log *logger.Logger) slog.Handler {
	return NewWebsocketLogHandler(log, &f.wsLogMuted, "Subscription acknowledged successfully")
}

func (h *websocketLogHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *websocketLogHandler) Handle(_ context.Context, r slog.Record) error {
	if (h.muted != nil && h.muted.Load()) || h.drop[r.Message] {
		return nil
	}
	attrs := map[string]slog.Value{}
	order := []string{}
	add := func(a slog.Attr) bool {
		if a.Key == "component" {
			return true
		}
		if _, ok := attrs[a.Key]; !ok {
			order = append(order, a.Key)
		}
		attrs[a.Key] = a.Value.Resolve()
		return true
	}
	for _, a := range h.attrs {
		add(a)
	}
	r.Attrs(add)

	message := websocketMessage(r.Message, attrs, order)
	if r.Message == "Subscribing to pattern" {
		// v2.59.0 logged the channels before it subscribed
		h.log.Infof("Listening to subscriptions: %s", attrs["value"].String())
	}
	switch {
	case r.Level >= slog.LevelError:
		h.log.Error(message)
	case r.Level >= slog.LevelWarn:
		h.log.Warn(message)
	case r.Level >= slog.LevelInfo:
		h.log.Info(message)
	default:
		h.log.Debug(message)
	}
	return nil
}

func (h *websocketLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &websocketLogHandler{log: h.log, attrs: append(append([]slog.Attr(nil), h.attrs...), attrs...), muted: h.muted, drop: h.drop}
}

func (h *websocketLogHandler) WithGroup(string) slog.Handler { return h }

// websocketMessage returns the text of v2.59.0 for a record of the SDK
func websocketMessage(msg string, attrs map[string]slog.Value, order []string) string {
	str := func(key string) string {
		v, ok := attrs[key]
		if !ok {
			return ""
		}
		switch v.Kind() {
		case slog.KindAny:
			switch x := v.Any().(type) {
			case []byte:
				return string(x)
			case []string:
				return strings.Join(x, ",")
			}
		case slog.KindTime:
			return v.Time().String()
		}
		return v.String()
	}

	switch msg {
	// realtime
	case "Establishing connection":
		return "Establishing connection to " + str("to")
	case "Waiting for message":
		return "Waiting for message: id=" + str("id")
	case "ws (send)", "ws (recv)":
		return fmt.Sprintf("%s: %s : %s", msg, str("channel"), str("text"))
	case "Pending messages":
		return "Pending messages ids: " + str("id")
	case "Removing message from pending requests":
		return "Removing message from pending requests: " + str("id")
	case "Received message":
		if _, ok := attrs["id"]; ok {
			return "Received message " + str("id")
		}
	case "Resending /meta/connect heartbeat after a small delay":
		return fmt.Sprintf("Resending /meta/connect heartbeat with %s ms delay", str("delay_ms"))
	case "Scheduling sending of new handshake to server with a small delay":
		return fmt.Sprintf("Scheduling sending of new handshake to server with %s ms delay", str("delay_ms"))
	case "Subscribing to pattern":
		return "Subscribing to " + str("value")
	case "Unsubscribing to pattern":
		return "unsubscribing to " + str("value")
	case "Successfully subscribed to channel", "Failed to subscribe to channel", "Successfully unsubscribed to channel":
		return msg + " " + str("value")
	case "ws ReadJSON":
		return fmt.Sprintf("wc ReadJSON: error=%s, message=%s", str("err"), str("message"))
	case "Failed to disconnect. %s":
		return "Failed to disconnect. " + str("err")
	case "Failed to get advice from server":
		return "Failed to get advice from server. " + str("err")
	case "Failed to send disconnect to server":
		return "Failed to send disconnect to server: " + str("err")
	case "Could not marshal message for sending", "Failed to send JSON message":
		return msg + ". " + str("err")

	// notification2
	case "Existing token":
		return fmt.Sprintf("Existing token: alg=%s, valid=%s, expired=%s, issuedAt: %s, expiresAt: %s, subscription=%s, subscriber=%s, shared=%s, tenant=%s",
			str("alg"), str("valid"), str("expired"), str("issuedAt"), str("expiresAt"), str("subscription"), str("subscriber"), str("shared"), str("tenant"))
	}

	if len(order) == 0 {
		return msg
	}
	if len(order) == 1 && order[0] == "err" {
		return msg + ". " + str("err")
	}
	details := make([]string, 0, len(order))
	for _, key := range order {
		details = append(details, key+": "+str(key))
	}
	return msg + ". " + strings.Join(details, ", ")
}
