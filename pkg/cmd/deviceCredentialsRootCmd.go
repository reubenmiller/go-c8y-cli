package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type DeviceCredentialsCmd struct {
	*subcommand.SubCommand
}

func NewDeviceCredentialsRootCmd() *DeviceCredentialsCmd {
	ccmd := &DeviceCredentialsCmd{}

	cmd := &cobra.Command{
		Use:   "deviceCredentials",
		Short: "Cumulocity device credentials",
		Long:  `REST endpoint to interact with Cumulocity device credentials api`,
	}

	// Subcommands
	cmd.AddCommand(NewGetNewDeviceRequestCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetNewDeviceRequestCmd().GetCommand())
	cmd.AddCommand(NewRegisterNewDeviceCmd().GetCommand())
	cmd.AddCommand(NewApproveNewDeviceRequestCmd().GetCommand())
	cmd.AddCommand(NewDeleteNewDeviceRequestCmd().GetCommand())
	cmd.AddCommand(NewRequestDeviceCredentialsCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
