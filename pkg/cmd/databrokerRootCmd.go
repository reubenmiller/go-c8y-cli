package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type DatabrokerCmd struct {
	*subcommand.SubCommand
}

func NewDatabrokerRootCmd() *DatabrokerCmd {
	ccmd := &DatabrokerCmd{}

	cmd := &cobra.Command{
		Use:   "databroker",
		Short: "Cumulocity databroker",
		Long:  `REST endpoint to interact with Cumulocity databroker`,
	}

	// Subcommands
	cmd.AddCommand(NewGetDataBrokerConnectorCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetDataBrokerCmd().GetCommand())
	cmd.AddCommand(NewUpdateDataBrokerCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
