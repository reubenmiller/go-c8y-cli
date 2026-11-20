// Package devicegroups wires the `c8y devicegroups` command and its subcommands.
// All subcommands are hand-written v2 c8ystream commands (backed by the go-c8y v2
// DeviceGroups service, and ManagedObjects child-reference services for the
// deprecated child-asset operations), so this group command is itself
// hand-written rather than generated.
package devicegroups

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdDevicegroups struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdDevicegroups {
	ccmd := &SubCmdDevicegroups{}

	cmd := &cobra.Command{
		Use:   "devicegroups",
		Short: "Cumulocity device groups",
		Long:  `REST endpoint to interact with Cumulocity device groups`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewAssignDeviceCmd(f).GetCommand())
	cmd.AddCommand(NewAssignGroupCmd(f).GetCommand())
	cmd.AddCommand(NewUnassignDeviceCmd(f).GetCommand())
	cmd.AddCommand(NewUnassignGroupCmd(f).GetCommand())
	cmd.AddCommand(NewListAssetsCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
