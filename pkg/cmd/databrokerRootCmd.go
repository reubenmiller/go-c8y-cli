package cmd

import (
	"github.com/spf13/cobra"
)

type databrokerCmd struct {
	*baseCmd
}

func newDatabrokerRootCmd() *databrokerCmd {
	ccmd := &databrokerCmd{}

	cmd := &cobra.Command{
		Use:   "databroker",
		Short: "Cumulocity databroker",
		Long:  `REST endpoint to interact with Cumulocity databroker`,
	}

	// Subcommands
	cmd.AddCommand(newGetDataBrokerConnectorCollectionCmd().getCommand())
	cmd.AddCommand(newGetDataBrokerCmd().getCommand())
	cmd.AddCommand(newUpdateDataBrokerCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
