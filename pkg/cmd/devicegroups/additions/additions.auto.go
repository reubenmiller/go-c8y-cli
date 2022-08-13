package additions

import (
	cmdAssign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/additions/assign"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/additions/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/additions/list"
	cmdUnassign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/additions/unassign"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAdditions struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdAdditions {
	ccmd := &SubCmdAdditions{}

	cmd := &cobra.Command{
		Use:   "additions",
		Short: "Cumulocity device groups additions",
		Long:  `Manage child additions in devicegroups`,
	}

	// Subcommands
	cmd.AddCommand(cmdAssign.NewAssignCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassign.NewUnassignCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
