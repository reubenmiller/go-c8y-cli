package assets

import (
	cmdAssign "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/assets/assign"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/assets/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/assets/list"
	cmdUnassign "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/assets/unassign"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAssets struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdAssets {
	ccmd := &SubCmdAssets{}

	cmd := &cobra.Command{
		Use:   "assets",
		Short: "Cumulocity inventory assets",
		Long:  `REST endpoint to interact with Cumulocity managed objects`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdAssign.NewAssignCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassign.NewUnassignCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
