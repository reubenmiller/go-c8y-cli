package children

import (
	cmdAssign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/children/assign"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/children/create"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/children/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/children/list"
	cmdUnassign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/children/unassign"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdChildren struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdChildren {
	ccmd := &SubCmdChildren{}

	cmd := &cobra.Command{
		Use:   "children",
		Short: "Cumulocity managed object child references",
		Long:  `Manage child entities (assets, additions and device) for managed objects`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdAssign.NewAssignCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassign.NewUnassignCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
