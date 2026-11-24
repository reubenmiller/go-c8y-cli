// Package c8ywait polls Cumulocity until something reaches a desired state.
// It is used by the wait and assert commands; each command only says how to
// fetch the current state and whether it is the one wanted.
package c8ywait

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
)

// DefaultTimeout is used when Options.Timeout is 0.
const DefaultTimeout = 300 * time.Second

// Options control how long and how often the state is checked.
type Options struct {
	// Interval between two checks. 0 means one second.
	Interval time.Duration
	// Timeout for the whole wait. 0 means DefaultTimeout, a negative value
	// waits forever.
	Timeout time.Duration
	// Attempts is the most number of checks. 0 or less is unlimited.
	Attempts int64
}

// State is the result of one check.
type State struct {
	// Doc is the current state of the object, shown to the user when the
	// wait ends. It may be empty.
	Doc jsondoc.JSONDoc
	// Done ends the wait.
	Done bool
	// Err says why the state is not the wanted one. When Done is set too, the
	// wait ends with this error (e.g. an operation that failed while waiting
	// for it to succeed).
	Err error
}

// Check fetches the current state. An error ends the wait at once (e.g. the
// object does not exist or the request failed).
type Check func(ctx context.Context) (State, error)

// Until checks the state until it is done, the attempts are used up, the
// timeout passes or the user presses Ctrl-C. It returns the last document
// and the error to report, if any.
//
// In a dry run it checks once, so the request is shown, and returns no
// document.
func Until(ctx context.Context, opts Options, check Check) (jsondoc.JSONDoc, error) {
	if apiv2.IsDryRun(ctx) {
		// Nothing is sent, so the state never changes: check once, which
		// shows the request, and report whether that (empty) state passes.
		state, err := check(ctx)
		if err != nil {
			return jsondoc.Empty(), err
		}
		return jsondoc.Empty(), state.Err
	}

	interval := opts.Interval
	if interval <= 0 {
		interval = time.Second
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	waitCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	runCtx, stop := signal.NotifyContext(waitCtx, os.Interrupt)
	defer stop()

	last := jsondoc.Empty()
	var lastErr error
	ended := func() error {
		switch {
		case ctx.Err() != nil:
			return ctx.Err()
		case waitCtx.Err() != nil:
			msg := fmt.Sprintf("Timeout after %v", timeout)
			if lastErr != nil {
				msg = fmt.Sprintf("%s: %s", msg, lastErr)
			}
			return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitTimeout, msg)
		default:
			return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitCancel, "User cancelled command")
		}
	}

	for attempt := int64(1); ; attempt++ {
		state, err := check(runCtx)
		if err != nil {
			if runCtx.Err() != nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
				return last, ended()
			}
			return last, err
		}
		if !isEmpty(state.Doc) {
			last = state.Doc
		}
		if state.Done {
			return last, state.Err
		}
		lastErr = state.Err
		if opts.Attempts > 0 && attempt >= opts.Attempts {
			return last, lastErr
		}

		timer := time.NewTimer(interval)
		select {
		case <-timer.C:
		case <-runCtx.Done():
			timer.Stop()
			return last, ended()
		}
	}
}

// CompareCount reports whether count is within the minimum and maximum. A
// negative limit is not checked.
func CompareCount(count, minimum, maximum int64) bool {
	switch {
	case minimum > -1 && maximum > -1:
		return count >= minimum && count <= maximum
	case minimum > -1:
		return count >= minimum
	case maximum > -1:
		return count <= maximum
	}
	return false
}

// Seq yields doc (when it is not empty) and then err (when set), the shape a
// c8ystream Call returns.
func Seq(doc jsondoc.JSONDoc, err error) func(yield func(jsondoc.JSONDoc, error) bool) {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		if !isEmpty(doc) {
			if !yield(doc, nil) {
				return
			}
		}
		if err != nil {
			yield(jsondoc.Empty(), err)
		}
	}
}

// isEmpty reports whether doc holds no document (a zero JSONDoc or {}).
func isEmpty(doc jsondoc.JSONDoc) bool {
	b := doc.Bytes()
	return len(b) == 0 || string(b) == "{}" || string(b) == "null"
}
