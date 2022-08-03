package children

import (
	cmdAssign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/children/assign"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/children/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/children/list"
	cmdUnassign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/children/unassign"
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
		Short: "Cumulocity device children",
		Long:  `Managed device children`,
	}

	// Subcommands
	cmd.AddCommand(cmdAssign.NewAssignCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassign.NewUnassignCmd(f).GetCommand())
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
