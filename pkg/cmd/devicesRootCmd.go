package cmd

import (
	"github.com/spf13/cobra"
)

type devicesCmd struct {
	*baseCmd
}

func newDevicesRootCmd() *devicesCmd {
	ccmd := &devicesCmd{}

	cmd := &cobra.Command{
		Use:   "devices",
		Short: "Cumulocity devices",
		Long:  `REST endpoint to interact with Cumulocity devices`,
	}

	// Subcommands
	cmd.AddCommand(newGetDeviceCmd().getCommand())
	cmd.AddCommand(newUpdateDeviceCmd().getCommand())
	cmd.AddCommand(newDeleteDeviceCmd().getCommand())
	cmd.AddCommand(newCreateDeviceCmd().getCommand())
	cmd.AddCommand(newGetSupportedMeasurementsCmd().getCommand())
	cmd.AddCommand(newGetSupportedSeriesCmd().getCommand())
	cmd.AddCommand(newGetSupportedOperationsCmd().getCommand())
	cmd.AddCommand(newSetDeviceRequiredAvailabilityCmd().getCommand())
	cmd.AddCommand(newGetDeviceGroupCmd().getCommand())
	cmd.AddCommand(newUpdateDeviceGroupCmd().getCommand())
	cmd.AddCommand(newDeleteDeviceGroupCmd().getCommand())
	cmd.AddCommand(newCreateDeviceGroupCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
