// v2-based measurements deleteCollection: deletes the measurements matching the
// given filter (device/date/type/fragmentType) in one call via
// Measurements.DeleteList. The device flag drives iteration (pipe or --device).
// Success yields no output.
package deletecollection

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/measurements"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteCollectionCmd command
type DeleteCollectionCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteCollectionCmd creates a command to Delete measurement collection
func NewDeleteCollectionCmd(f *cmdutil.Factory) *DeleteCollectionCmd {
	ccmd := &DeleteCollectionCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "deleteCollection",
		Short: "Delete measurement collection",
		Long:  `Delete measurements using a filter`,
		Example: heredoc.Doc(`
$ c8y measurements deleteCollection --device 12345
Delete measurements for a device

$ c8y measurements deleteCollection --device 12345 --dateTo "-10d"
Delete measurements older than 10 days for a device

$ c8y measurements deleteCollection --device 12345 --dateTo "-10d" --fragmentType lmp
Delete measurements with a given fragment and older than 10 days for a device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device ID (accepts pipeline)")
	cmd.Flags().String("type", "", "Measurement type.")
	cmd.Flags().String("fragmentType", "", "Fragment name from measurement")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of measurement occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of measurement occurrence.")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("device", "source.id", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Remove-MeasurementCollection"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
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

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := measurements.DeleteListOptions{
			DateFrom:     in.TimeValue("dateFrom"),
			DateTo:       in.TimeValue("dateTo"),
			Type:         in.String("type"),
			FragmentType: in.String("fragmentType"),
		}
		if device := in.String("device"); device != "" {
			opt.Source = managedobjects.DeviceRef(c8ystream.NameOrID(device))
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Measurements.DeleteList(ctx, opt)
			})
		}, nil
	})
}
