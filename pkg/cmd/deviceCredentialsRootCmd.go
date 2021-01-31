package cmd

import (
	"github.com/spf13/cobra"
)

type DeviceCredentialsCmd struct {
	*baseCmd
}

func NewDeviceCredentialsRootCmd() *DeviceCredentialsCmd {
	ccmd := &DeviceCredentialsCmd{}

	cmd := &cobra.Command{
		Use:   "deviceCredentials",
		Short: "Cumulocity device credentials",
		Long:  `REST endpoint to interact with Cumulocity device credentials api`,
	}

	// Subcommands
	cmd.AddCommand(NewGetNewDeviceRequestCollectionCmd().getCommand())
	cmd.AddCommand(NewGetNewDeviceRequestCmd().getCommand())
	cmd.AddCommand(NewRegisterNewDeviceCmd().getCommand())
	cmd.AddCommand(NewApproveNewDeviceRequestCmd().getCommand())
	cmd.AddCommand(NewDeleteNewDeviceRequestCmd().getCommand())
	cmd.AddCommand(NewRequestDeviceCredentialsCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
