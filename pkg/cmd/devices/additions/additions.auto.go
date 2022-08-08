package additions

import (
	cmdAssign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/additions/assign"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/additions/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/additions/list"
	cmdUnassign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/additions/unassign"
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
		Short: "Cumulocity device additions",
		Long:  `REST endpoint to interact with Cumulocity device additions`,
	}

	// Subcommands
	cmd.AddCommand(cmdAssign.NewAssignCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassign.NewUnassignCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
