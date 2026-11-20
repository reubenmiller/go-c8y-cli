// Package devices wires the `c8y devices` command and its flat subcommands. All
// subcommands are hand-written v2 c8ystream commands backed by the go-c8y v2
// Devices / ManagedObjects services, so this group command is itself
// hand-written rather than generated. The children/services/statistics/user/
// availability subgroups are attached by pkg/cmd/root.
package devices

import (
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
		Short: "Cumulocity devices",
		Long:  `REST endpoint to interact with Cumulocity devices`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewGetSupportedMeasurementsCmd(f).GetCommand())
	cmd.AddCommand(NewGetSupportedSeriesCmd(f).GetCommand())
	cmd.AddCommand(NewAssignChildCmd(f).GetCommand())
	cmd.AddCommand(NewUnassignChildCmd(f).GetCommand())
	cmd.AddCommand(NewListChildrenCmd(f).GetCommand())
	cmd.AddCommand(NewGetChildCmd(f).GetCommand())
	cmd.AddCommand(NewListAssetsCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
