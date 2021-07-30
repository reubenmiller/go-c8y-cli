package additions

import (
	cmdAssign "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/additions/assign"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/additions/create"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/additions/list"
	cmdUnassign "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/additions/unassign"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAdditions struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdAdditions {
	ccmd := &SubCmdAdditions{}

	cmd := &cobra.Command{
		Use:   "additions",
		Short: "Cumulocity inventory additions",
		Long:  `REST endpoint to interact with Cumulocity managed objects`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdAssign.NewAssignCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassign.NewUnassignCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
