package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenantoptions/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenantoptions/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenantoptions/get"
	cmdGetForCategory "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenantoptions/getforcategory"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenantoptions/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenantoptions/update"
	cmdUpdateBulk "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenantoptions/updatebulk"
	cmdUpdateEdit "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenantoptions/updateedit"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdTenantoptions struct {
	*subcommand.SubCommand
}

func NewSubCmdTenantoptions(f *cmdutil.Factory) *SubCmdTenantoptions {
	ccmd := &SubCmdTenantoptions{}

	cmd := &cobra.Command{
		Use:   "tenantoptions",
		Short: "Cumulocity tenantOptions",
		Long: `<
REST endpoint to interact with Cumulocity tenantOptions
Options are category-key-value tuples, storing tenant configuration. Some categories of options allow creation of new one, other are limited to predefined set of keys.

Any option of any tenant can be defined as "non-editable" by "management" tenant. Afterwards, any PUT or DELETE requests made on that option by the owner tenant, will result in 403 error (Unauthorized).`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdateBulk.NewUpdateBulkCmd(f).GetCommand())
	cmd.AddCommand(cmdGetForCategory.NewGetForCategoryCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdateEdit.NewUpdateEditCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
