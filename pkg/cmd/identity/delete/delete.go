// v2-based identity delete: the name flag (the external id) drives iteration
// (pipe or --name); each external identity is removed by type + external id via
// Identity.Delete. Deleting an external id does not delete the managed object.
package delete

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/identity"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteCmd command
type DeleteCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteCmd creates a command to Delete external identity
func NewDeleteCmd(f *cmdutil.Factory) *DeleteCmd {
	ccmd := &DeleteCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete external identity",
		Long:  `Delete an existing external id. This does not delete the device managed object`,
		Example: heredoc.Doc(`
$ c8y identity delete --type test --name myserialnumber
Delete external identity

$ c8y devices list | c8y identity list --filter 'type eq c8y_Serial' | c8y identity delete --type c8y_Serial
Delete a specific external identity type (via pipeline)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("type", "c8y_Serial", "External identity type")
	cmd.Flags().String("name", "", "External identity name (required) (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("name", "name", true, "externalId", "name", "id"),
		flags.WithPowershellName("Remove-ExternalId"),
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

	if err := r.InputFlag("name"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opts := identity.IdentityOptions{
			Type:       in.String("type"),
			ExternalID: in.String("name"),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Identity.Delete(ctx, opts)
			})
		}, nil
	})
}
