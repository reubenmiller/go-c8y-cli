package devices

import (
	cmdAssign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/devices/assign"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/devices/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/devices/list"
	cmdUnassign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/devices/unassign"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdDevices struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdDevices {
	ccmd := &SubCmdDevices{}

	cmd := &cobra.Command{
		Use:   "devices",
		Short: "Cumulocity device groups devices",
		Long:  `Manage child devices in devicegroups`,
	}

	// Subcommands
	cmd.AddCommand(cmdAssign.NewAssignCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassign.NewUnassignCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
