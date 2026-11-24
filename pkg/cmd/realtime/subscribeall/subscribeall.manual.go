package subscribeall

import (
	"context"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/realtime"
	"github.com/spf13/cobra"
)

type CmdSubscribeAll struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdSubscribeAll(f *cmdutil.Factory) *CmdSubscribeAll {
	ccmd := &CmdSubscribeAll{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "subscribeAll",
		Short: "Subscribe to all realtime notifications",
		Long:  `Subscribe to all realtime notifications`,
		Example: heredoc.Doc(`
$ c8y realtime subscribeAll --device 12345 --duration 90s

Subscribe to all notifications (alarms/events/operations etc.) for device 12345 for 90 seconds
		`),
		RunE: ccmd.RunE,
	}

	// Flags
	cmd.Flags().StringSlice("device", []string{""}, "Device ID or name (accepts pipeline). Without a device, all devices are used")
	cmd.Flags().String("duration", "30s", "Duration to subscribe for. i.e. 30s, 1m")
	cmd.Flags().Int64("count", 0, "Max number of realtime notifications to wait for")
	cmd.Flags().StringSlice("actionTypes", nil, "Filter by realtime action types, i.e. CREATE,UPDATE,DELETE")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("actionTypes", "CREATE", "UPDATE", "DELETE"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "device", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE subscribes to the alarms, events, measurements and operations of the
// device (or of all devices with no --device) on one realtime stream.
func (n *CmdSubscribeAll) RunE(cmd *cobra.Command, args []string) error {
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
		return func(ctx context.Context) any {
			id := "*"
			if device != "" {
				resolved, err := client.ManagedObjects.ResolveID(apiv2.ResolutionContext(ctx), c8ystream.NameOrID(device), nil)
				if err != nil {
					return err
				}
				id = resolved
			}
			return c8ystream.Subscribed(client.Realtime().SubscribeStream(ctx,
				realtime.Alarms(id),
				realtime.Events(id),
				realtime.Measurements(id),
				realtime.Operations(id),
			), opts)
		}, nil
	})
}
