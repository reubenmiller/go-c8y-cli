// v2-based inventory role list: lists all inventory roles via
// InventoryRoles.ListAll. There is no pipeline input, so the command runs once.

package users

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	inventoryroles "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/userroles/inventory"
	"github.com/spf13/cobra"
)

// ListInventoryRolesCmd command
type ListInventoryRolesCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListInventoryRolesCmd creates a command to Get inventory role collection
func NewListInventoryRolesCmd(f *cmdutil.Factory) *ListInventoryRolesCmd {
	ccmd := &ListInventoryRolesCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "listInventoryRoles",
		Short: "Get inventory role collection",
		Long:  `Get a list of inventory roles`,
		Example: heredoc.Doc(`
$ c8y users listInventoryRoles
Get list of inventory roles
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithCollectionProperty("roles"),
		flags.WithPowershellName("Get-InventoryRoleCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.inventoryrolecollection+json", "application/vnd.com.nsn.cumulocity.inventoryrole+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListInventoryRolesCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	paging, err := r.Pagination()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := inventoryroles.ListOptions{}
		opt.PaginationOptions = paging
		return func(ctx context.Context) any {
			return client.InventoryRoles.ListAll(ctx, opt)
		}, nil
	})
}
