package cmd

import (
	"github.com/spf13/cobra"
)

type DevicesCmd struct {
	*baseCmd
}

func NewDevicesRootCmd() *DevicesCmd {
	ccmd := &DevicesCmd{}

	cmd := &cobra.Command{
		Use:   "devices",
		Short: "Cumulocity devices",
		Long:  `REST endpoint to interact with Cumulocity devices`,
	}

	// Subcommands
	cmd.AddCommand(NewGetDeviceCmd().getCommand())
	cmd.AddCommand(NewUpdateDeviceCmd().getCommand())
	cmd.AddCommand(NewDeleteDeviceCmd().getCommand())
	cmd.AddCommand(NewCreateDeviceCmd().getCommand())
	cmd.AddCommand(NewGetSupportedMeasurementsCmd().getCommand())
	cmd.AddCommand(NewGetSupportedSeriesCmd().getCommand())
	cmd.AddCommand(NewGetSupportedOperationsCmd().getCommand())
	cmd.AddCommand(NewSetDeviceRequiredAvailabilityCmd().getCommand())
	cmd.AddCommand(NewGetDeviceGroupCmd().getCommand())
	cmd.AddCommand(NewUpdateDeviceGroupCmd().getCommand())
	cmd.AddCommand(NewDeleteDeviceGroupCmd().getCommand())
	cmd.AddCommand(NewCreateDeviceGroupCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
