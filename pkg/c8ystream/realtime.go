package c8ystream

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/realtime"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
)

// StreamOptions reads the common realtime flags of a subscribe command:
// --duration (default unit seconds), --count and --actionTypes.
func (r *Runner) StreamOptions() (realtime.StreamOptions, error) {
	// a command which streams logs as v2.59.0 did: no jobs, no response details
	r.Factory.SetStreaming()
	opts := realtime.StreamOptions{}
	if r.Cmd.Flags().Lookup("duration") != nil {
		d, err := flags.GetDurationFlag(r.Cmd, "duration", true, time.Second)
		if err != nil {
			return opts, err
		}
		opts.Timeout = d
	}
	if f := r.Cmd.Flags().Lookup("count"); f != nil {
		if n, err := r.Cmd.Flags().GetInt64("count"); err == nil {
			opts.MaxMessages = n
		}
	}
	if f := r.Cmd.Flags().Lookup("actionTypes"); f != nil {
		if v, err := r.Cmd.Flags().GetStringSlice("actionTypes"); err == nil {
			opts.Actions = v
		}
	}
	return opts, nil
}

// Subscribed applies the stream options to a realtime subscription result, so
// a subscribe command's Call can return it like any other SDK result.
func Subscribed[T any](res op.Result[*realtime.Stream[T]], opts realtime.StreamOptions) op.Result[*realtime.Stream[T]] {
	if res.Err == nil && res.Data != nil {
		res.Data.WithOptions(opts)
	}
	return res
}

// Subscribe runs the subscribe call of a realtime command: it applies the stream
// options (see Subscribed) and, as v2.59.0 did:
//
//   - Ctrl-C cancels the connection while it is made (the handshake of the
//     websocket and of the realtime protocol), instead of waiting for it;
//   - the end of the stream is logged ("Duration has expired. Stopping realtime
//     client", or "Stopping realtime client" for Ctrl-C), and nothing of the
//     clean up of the connection is logged after it.
//
// A result which is not a stream (an error, a dry run) is returned as it is.
func Subscribe[T, A any](r *Runner, ctx context.Context, opts realtime.StreamOptions, subscribe func(context.Context, A) op.Result[*realtime.Stream[T]], arg A) any {
	return subscribeStream(r, ctx, opts, func(ctx context.Context) op.Result[*realtime.Stream[T]] {
		return subscribe(ctx, arg)
	})
}

// SubscribePatterns is Subscribe for a subscription to channel patterns.
func SubscribePatterns[T any](r *Runner, ctx context.Context, opts realtime.StreamOptions, subscribe func(context.Context, ...string) op.Result[*realtime.Stream[T]], patterns ...string) any {
	return subscribeStream(r, ctx, opts, func(ctx context.Context) op.Result[*realtime.Stream[T]] {
		return subscribe(ctx, patterns...)
	})
}

// realtimeConnectError gives the error of a realtime websocket which could not
// be opened (connection refused, a rejected upgrade, ...) the text and the exit
// code of v2.59.0: a command error "could not create realtime client. <error>"
// with exit code 100. Any other error is returned as it is.
func realtimeConnectError(err error) error {
	var connErr *realtime.ConnectError
	if !errors.As(err, &connErr) || connErr.Err == nil {
		return err
	}
	return cmderrors.CommandError{
		Message:   "could not create realtime client. " + connErr.Err.Error(),
		ErrorType: cmderrors.ErrTypeCommand,
		ExitCode:  cmderrors.ExitError,
	}
}

func subscribeStream[T any](r *Runner, ctx context.Context, opts realtime.StreamOptions, subscribe func(context.Context) op.Result[*realtime.Stream[T]]) any {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// The stream ends when streamCtx ends: after the duration, or with Ctrl-C (also
	// while the connection is made, which would otherwise wait for its timeout).
	streamCtx, cancel := context.WithCancel(ctx)
	log, _ := r.Factory.Logger()
	var connected atomic.Bool
	stop := func(message string) {
		if log != nil {
			log.Info(message)
		}
		r.Factory.MuteWebsocketLog()
		cancel()
	}
	done := make(chan struct{})
	finish := func() {
		close(done)
		signal.Stop(interrupt)
	}
	go func() {
		select {
		case <-interrupt:
			if connected.Load() {
				stop("Stopping realtime client")
			} else {
				cancel()
			}
		case <-done:
		case <-streamCtx.Done():
		}
	}()

	res := subscribe(streamCtx)
	if res.Err != nil || res.Data == nil {
		finish()
		if errors.Is(res.Err, context.Canceled) && ctx.Err() == nil {
			// Ctrl-C while connecting
			res.Err = cmderrors.NewUserErrorWithExitCode(cmderrors.ExitCancel, "Interrupted")
		}
		res.Err = realtimeConnectError(res.Err)
		return Subscribed(res, opts)
	}
	connected.Store(true)
	// the duration is applied here, to log it
	duration := opts.Timeout
	opts.Timeout = 0
	res = Subscribed(res, opts)

	return output.Seq(func(yield func(jsondoc.JSONDoc, error) bool) {
		defer finish()
		if duration > 0 {
			timer := time.AfterFunc(duration, func() { stop("Duration has expired. Stopping realtime client") })
			defer timer.Stop()
		}
		var count int64
		for doc, err := range res.Documents() {
			if err != nil && errors.Is(err, context.Canceled) && streamCtx.Err() != nil && ctx.Err() == nil {
				// stopped by the duration or by Ctrl-C
				return
			}
			if err == nil && opts.MaxMessages > 0 && count+1 >= opts.MaxMessages {
				// the stream ends with this message: nothing of the clean up is logged
				r.Factory.MuteWebsocketLog()
			}
			if !yield(doc, err) {
				return
			}
			if err == nil {
				count++
			}
		}
	})
}
