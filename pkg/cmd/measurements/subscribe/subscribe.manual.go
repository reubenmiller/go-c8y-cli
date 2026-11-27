package subscribe

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/spf13/cobra"
)

type CmdSubscribe struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdSubscribe(f *cmdutil.Factory) *CmdSubscribe {
	ccmd := &CmdSubscribe{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to realtime measurements",
		Long:  `Subscribe to realtime measurements`,
		Example: heredoc.Doc(`
$ c8y measurements subscribe --device 12345
Subscribe to measurements (in realtime) for device 12345

$ c8y measurements subscribe --device 12345 --duration 30s
Subscribe to measurements (in realtime) for device 12345 for 30 seconds

$ c8y measurements subscribe --count 10
Subscribe to measurements (in realtime) for all devices, and stop after receiving 10 measurements
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device ID or name (accepts pipeline). Without a device, all devices are used")
	cmd.Flags().String("duration", "30s", "Duration to subscribe for. i.e. 30s, 1m")
	cmd.Flags().Int64("count", 0, "Max number of realtime notifications to wait for")
	cmd.Flags().StringSlice("actionTypes", nil, "Filter by realtime action types, i.e. CREATE,UPDATE,DELETE")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", ccmd.factory.ClientV2),
		completion.WithValidateSet("actionTypes", "CREATE", "UPDATE", "DELETE"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "device", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE subscribes to the realtime measurements of the device (or of all devices
// with no --device) until --duration passes or --count notifications arrived.
func (n *CmdSubscribe) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	if err := r.InputFlag("device"); err != nil {
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

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		device := in.String("device")
		if device == "" {
			device = "*"
		}
		return func(ctx context.Context) any {
			return c8ystream.Subscribe(r, ctx, opts, client.Measurements.SubscribeStream, device)
		}, nil
	})
}
