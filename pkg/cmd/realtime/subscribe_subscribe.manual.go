package realtime

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/realtime"
	"github.com/spf13/cobra"
)

type CmdSubscribe struct {
	flagChannel string

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdSubscribe(f *cmdutil.Factory) *CmdSubscribe {
	ccmd := &CmdSubscribe{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to realtime notifications",
		Long:  `Subscribe to realtime notifications`,
		Example: heredoc.Doc(`
$ c8y realtime subscribe --channel "/measurements/*" --duration 90s

Subscribe to all measurements for 90 seconds
		`),
		RunE: ccmd.RunE,
	}

	// Flags
	cmd.Flags().StringVar(&ccmd.flagChannel, "channel", "", "Channel name i.e. \"/measurements/12345\" or \"/measurements/*\"")
	cmd.Flags().String("duration", "30s", "Duration to subscribe for. i.e. 30s, 1m")
	cmd.Flags().Int64("count", 0, "Max number of realtime notifications to wait for")
	cmd.Flags().StringSlice("actionTypes", nil, "Filter by realtime action types, i.e. CREATE,UPDATE,DELETE")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("actionTypes", "CREATE", "UPDATE", "DELETE"),
		completion.WithValidateSet(
			"channel",
			realtime.Alarms("*"),
			realtime.AlarmsWithChildren("*"),
			realtime.Events("*"),
			realtime.ManagedObjects("*"),
			realtime.Measurements("*"),
			realtime.Operations("*"),
		),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE subscribes to a realtime channel (e.g. /measurements/12345) until
// --duration passes or --count notifications arrived.
func (n *CmdSubscribe) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	client, err := r.Client()
	if err != nil {
		return err
	}
	opts, err := r.StreamOptions()
	if err != nil {
		return err
	}
	channel := n.flagChannel
	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		return func(ctx context.Context) any {
			return c8ystream.SubscribePatterns(r, ctx, opts, client.Realtime().SubscribeStream, channel)
		}, nil
	})
}
