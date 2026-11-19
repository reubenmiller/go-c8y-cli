// v2-based events deleteCollection: deletes the events matching the given filter
// (device/date/type/fragmentType) in one call via Events.DeleteList. The device
// flag drives iteration (pipe or --device). Success yields no output.
//
// The --revert flag is accepted for backwards compatibility but has no effect on
// a delete (the typed DeleteListOptions does not model it).
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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/events"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteCollectionCmd command
type DeleteCollectionCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteCollectionCmd creates a command to Delete event collection
func NewDeleteCollectionCmd(f *cmdutil.Factory) *DeleteCollectionCmd {
	ccmd := &DeleteCollectionCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "deleteCollection",
		Short: "Delete event collection",
		Long:  `Delete a collection of events by a given filter`,
		Example: heredoc.Doc(`
$ c8y events deleteCollection --type my_CustomType --dateFrom "-10d"
Remove events with type 'my_CustomType' that were created in the last 10 days

$ c8y events deleteCollection --device 12345
Remove events from a device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device ID (accepts pipeline)")
	cmd.Flags().String("type", "", "Event type.")
	cmd.Flags().String("fragmentType", "", "Fragment name from event.")
	cmd.Flags().String("createdFrom", "", "Start date or date and time of the event's creation (set by the platform during creation).")
	cmd.Flags().String("createdTo", "", "End date or date and time of the event's creation (set by the platform during creation).")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of event occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of event occurrence.")
	cmd.Flags().Bool("revert", false, "Return the newest instead of the oldest events. Must be used with dateFrom and dateTo parameters")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("device", "source.id", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Remove-EventCollection"),
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
		opt := events.DeleteListOptions{
			DateFrom:     in.TimeValue("dateFrom"),
			DateTo:       in.TimeValue("dateTo"),
			CreatedFrom:  in.TimeValue("createdFrom"),
			CreatedTo:    in.TimeValue("createdTo"),
			Type:         in.String("type"),
			FragmentType: in.String("fragmentType"),
		}
		if device := in.String("device"); device != "" {
			opt.Source = managedobjects.DeviceRef(c8ystream.NameOrID(device))
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Events.DeleteList(ctx, opt)
			})
		}, nil
	})
}
