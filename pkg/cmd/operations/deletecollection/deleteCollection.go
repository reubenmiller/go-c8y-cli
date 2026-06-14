// v2-based operations deleteCollection: deletes a set of operations matching the
// given filter (device/agent/date/status) in one call via Operations.DeleteList.
// The device flag drives iteration (pipe or --device), so a stream of devices
// deletes each device's matching operations. Success yields no output.
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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/operations"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/types"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteCollectionCmd command
type DeleteCollectionCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteCollectionCmd creates a command to Delete operation collection
func NewDeleteCollectionCmd(f *cmdutil.Factory) *DeleteCollectionCmd {
	ccmd := &DeleteCollectionCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "deleteCollection",
		Short: "Delete operation collection",
		Long:  `Delete a collection of operations by a given filter`,
		Example: heredoc.Doc(`
$ c8y operations deleteCollection --device 12345 --status PENDING
Remove all pending operations for a given device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("agent", []string{""}, "Agent ID")
	cmd.Flags().StringSlice("device", []string{""}, "Device ID (accepts pipeline)")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of operation.")
	cmd.Flags().String("dateTo", "", "End date or date and time of operation.")
	cmd.Flags().String("status", "", "Operation status, can be one of SUCCESSFUL, FAILED, EXECUTING or PENDING.")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "PENDING", "EXECUTING", "SUCCESSFUL", "FAILED"),
		completion.WithDevice("agent", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("device", "deviceId", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Remove-OperationCollection"),
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
		opt := operations.DeleteListOptions{
			DateFrom: in.TimeValue("dateFrom"),
			DateTo:   in.TimeValue("dateTo"),
			Status:   types.OperationStatus(in.String("status")),
		}
		if device := in.String("device"); device != "" {
			opt.DeviceID = managedobjects.DeviceRef(c8ystream.NameOrID(device))
		}
		if agents := in.StringSlice("agent"); len(agents) > 0 && agents[0] != "" {
			opt.AgentID = managedobjects.DeviceRef(c8ystream.NameOrID(agents[0]))
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Operations.DeleteList(ctx, opt)
			})
		}, nil
	})
}
