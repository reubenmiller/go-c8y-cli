// v2-based inventory delete: the id flag drives iteration (pipe or --id); each
// managed object (id or name, resolved internally) is deleted via
// ManagedObjects.Delete. Success yields no output.
package delete

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteCmd command
type DeleteCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteCmd creates a command to Delete managed object
func NewDeleteCmd(f *cmdutil.Factory) *DeleteCmd {
	ccmd := &DeleteCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete managed object",
		Long:  `Delete an existing managed object`,
		Example: heredoc.Doc(`
$ c8y inventory delete --id 12345
Delete a managed object

$ c8y inventory delete --id 12345 --cascade
Delete a managed object

$ c8y inventory delete --id 12345 --withDeviceUser
Delete a device and its related device user

$ c8y inventory delete --id 12345 --forceCascade
Delete a device and any related child assets, additions and/or devices
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "ManagedObject id (required) (accepts pipeline)")
	cmd.Flags().Bool("cascade", false, "When set to true and the managed object is a device or group, all the hierarchy will be deleted.")
	cmd.Flags().Bool("forceCascade", false, "When set to true all the hierarchy will be deleted without checking the type of managed object. It takes precedence over the parameter cascade.")
	cmd.Flags().Bool("withDeviceUser", false, "When set to true and the managed object is a device, it deletes the associated device user (credentials).")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("id", "id", true, "deviceId", "source.id", "managedObject.id"),
		flags.WithPowershellName("Remove-ManagedObject"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := managedobjects.DeleteOptions{
			Cascade:        in.Bool("cascade"),
			ForceCascade:   in.BoolPtrIfChanged("forceCascade"),
			WithDeviceUser: in.Bool("withDeviceUser"),
		}
		ref := c8ystream.NameOrID(in.String("id"))
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.ManagedObjects.Delete(ctx, ref, opt)
			})
		}, nil
	})
}
