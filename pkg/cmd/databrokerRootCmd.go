package cmd

import (
	"github.com/spf13/cobra"
)

type DatabrokerCmd struct {
	*baseCmd
}

func NewDatabrokerRootCmd() *DatabrokerCmd {
	ccmd := &DatabrokerCmd{}

	cmd := &cobra.Command{
		Use:   "databroker",
		Short: "Cumulocity databroker",
		Long:  `REST endpoint to interact with Cumulocity databroker`,
	}

	// Subcommands
	cmd.AddCommand(NewGetDataBrokerConnectorCollectionCmd().getCommand())
	cmd.AddCommand(NewGetDataBrokerCmd().getCommand())
	cmd.AddCommand(NewUpdateDataBrokerCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
