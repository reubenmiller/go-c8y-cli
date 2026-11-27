package c8ystream

import (
	"errors"
	"fmt"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
)

func TestRealtimeConnectError(t *testing.T) {
	cause := errors.New("websocket: bad handshake")
	err := realtimeConnectError(fmt.Errorf("realtime: failed to connect to ws://host/cep/realtime: %w", cause))
	var ce cmderrors.CommandError
	if !errors.As(err, &ce) {
		t.Fatalf("want a CommandError, got %T", err)
	}
	if ce.Message != "could not create realtime client. websocket: bad handshake" || ce.ExitCode != cmderrors.ExitError || ce.ErrorType != cmderrors.ErrTypeCommand {
		t.Errorf("unexpected error: %+v", ce)
	}

	other := errors.New("something else")
	if got := realtimeConnectError(other); got != other {
		t.Errorf("other errors are returned unchanged, got %v", got)
	}
	if got := realtimeConnectError(nil); got != nil {
		t.Errorf("nil stays nil, got %v", got)
	}
}

func TestHTTPCommandError(t *testing.T) {
	apiErr := serverErr("POST", "http://host/notification2/token", 403, `{"error":"security/Forbidden","message":"Access is denied"}`)
	ce, ok := HTTPCommandError(apiErr)
	if !ok {
		t.Fatal("want an HTTP command error")
	}
	if ce.ExitCode != cmderrors.ExitError || ce.ErrorType != cmderrors.ErrTypeCommand || ce.Message != "POST http://host/notification2/token: 403 security/Forbidden Access is denied" {
		t.Errorf("unexpected error: %+v", ce)
	}
	if _, ok := HTTPCommandError(errors.New("plain")); ok {
		t.Error("a plain error is not an HTTP error")
	}
}
