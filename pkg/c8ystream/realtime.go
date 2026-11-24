package c8ystream

import (
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/realtime"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
)

// StreamOptions reads the common realtime flags of a subscribe command:
// --duration (default unit seconds), --count and --actionTypes.
func (r *Runner) StreamOptions() (realtime.StreamOptions, error) {
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
