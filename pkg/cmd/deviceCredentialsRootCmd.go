package cmd

import (
	"github.com/spf13/cobra"
)

type deviceCredentialsCmd struct {
	*baseCmd
}

func newDeviceCredentialsRootCmd() *deviceCredentialsCmd {
	ccmd := &deviceCredentialsCmd{}

	cmd := &cobra.Command{
		Use:   "deviceCredentials",
		Short: "Cumulocity device credentials",
		Long:  `REST endpoint to interact with Cumulocity device credentials api`,
	}

	// Subcommands
	cmd.AddCommand(newGetNewDeviceRequestCollectionCmd().getCommand())
	cmd.AddCommand(newGetNewDeviceRequestCmd().getCommand())
	cmd.AddCommand(newRegisterNewDeviceCmd().getCommand())
	cmd.AddCommand(newApproveNewDeviceRequestCmd().getCommand())
	cmd.AddCommand(newDeleteNewDeviceRequestCmd().getCommand())
	cmd.AddCommand(newRequestDeviceCredentialsCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
