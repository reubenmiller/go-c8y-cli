package assets

import (
	cmdAssign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/assets/assign"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/assets/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/assets/list"
	cmdUnassign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/assets/unassign"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAssets struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdAssets {
	ccmd := &SubCmdAssets{}

	cmd := &cobra.Command{
		Use:   "assets",
		Short: "Cumulocity device assets",
		Long:  `REST endpoint to interact with Cumulocity device assets`,
	}

	// Subcommands
	cmd.AddCommand(cmdAssign.NewAssignCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassign.NewUnassignCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
