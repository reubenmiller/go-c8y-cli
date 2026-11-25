// Package tenantoptions wires the `c8y tenantoptions` command and its
// subcommands. All subcommands are hand-written v2 c8ystream commands, so this
// group command is itself hand-written rather than generated.
package tenantoptions

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdTenantoptions struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdTenantoptions {
	ccmd := &SubCmdTenantoptions{}

	cmd := &cobra.Command{
		Use:   "tenantoptions",
		Short: "Cumulocity tenantOptions",
		Long: `REST endpoint to interact with Cumulocity tenantOptions
Options are category-key-value tuples, storing tenant configuration. Some categories of options allow creation of new one, other are limited to predefined set of keys.

Any option of any tenant can be defined as "non-editable" by "management" tenant. Afterwards, any PUT or DELETE requests made on that option by the owner tenant, will result in 403 error (Unauthorized).
`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateBulkCmd(f).GetCommand())
	cmd.AddCommand(NewGetForCategoryCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateEditCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
