package children2

import (
	cmdAssign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/children2/assign"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/children2/create"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/children2/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/children2/list"
	cmdUnassign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/children2/unassign"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdChildren2 struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdChildren2 {
	ccmd := &SubCmdChildren2{}

	cmd := &cobra.Command{
		Use:   "children2",
		Short: "Cumulocity managed object child references",
		Long:  `Manage child entities (assets, additions and device) for devices`,
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
