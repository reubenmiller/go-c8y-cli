// v2-based get inventory role: the id flag drives iteration (pipe or --id); each
// numeric role id is fetched via InventoryRoles.Get. Lookup by name is not
// supported (the API has no by-name endpoint), so the id must be numeric.
package getinventoryrole

import (
	"context"
	"strconv"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetInventoryRoleCmd command
type GetInventoryRoleCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetInventoryRoleCmd creates a command to Get inventory role
func NewGetInventoryRoleCmd(f *cmdutil.Factory) *GetInventoryRoleCmd {
	ccmd := &GetInventoryRoleCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "getInventoryRole",
		Short: "Get inventory role",
		Long:  `Get a specific inventory role`,
		Example: heredoc.Doc(`
$ c8y users getInventoryRole --id 12345
Get an inventory role
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Role id. Note: lookup by name is not yet supported (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("Get-InventoryRole"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.inventoryrole+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetInventoryRoleCmd) RunE(cmd *cobra.Command, args []string) error {
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
		id, err := strconv.ParseInt(in.String("id"), 10, 64)
		if err != nil {
			return nil, cmderrors.NewUserError("invalid inventory role id (must be numeric): ", in.String("id"))
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.InventoryRoles.Get(ctx, id))
		}, nil
	})
}
