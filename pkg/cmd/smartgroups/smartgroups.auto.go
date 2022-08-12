package smartgroups

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/smartgroups/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/smartgroups/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/smartgroups/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/smartgroups/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/smartgroups/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdSmartgroups struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdSmartgroups {
	ccmd := &SubCmdSmartgroups{}

	cmd := &cobra.Command{
		Use:   "smartgroups",
		Short: "Cumulocity smart groups",
		Long:  `REST endpoint to interact with Cumulocity smart groups. A smart group is an inventory managed object and can also be managed via the Inventory api.`,
	}

	// Subcommands
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
