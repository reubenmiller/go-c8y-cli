package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type DevicesCmd struct {
	*subcommand.SubCommand
}

func NewDevicesRootCmd() *DevicesCmd {
	ccmd := &DevicesCmd{}

	cmd := &cobra.Command{
		Use:   "devices",
		Short: "Cumulocity devices",
		Long:  `REST endpoint to interact with Cumulocity devices`,
	}

	// Subcommands
	cmd.AddCommand(NewGetDeviceCmd().GetCommand())
	cmd.AddCommand(NewUpdateDeviceCmd().GetCommand())
	cmd.AddCommand(NewDeleteDeviceCmd().GetCommand())
	cmd.AddCommand(NewCreateDeviceCmd().GetCommand())
	cmd.AddCommand(NewGetSupportedMeasurementsCmd().GetCommand())
	cmd.AddCommand(NewGetSupportedSeriesCmd().GetCommand())
	cmd.AddCommand(NewGetSupportedOperationsCmd().GetCommand())
	cmd.AddCommand(NewSetDeviceRequiredAvailabilityCmd().GetCommand())
	cmd.AddCommand(NewGetDeviceGroupCmd().GetCommand())
	cmd.AddCommand(NewUpdateDeviceGroupCmd().GetCommand())
	cmd.AddCommand(NewDeleteDeviceGroupCmd().GetCommand())
	cmd.AddCommand(NewCreateDeviceGroupCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
