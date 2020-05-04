package cmd

import (
	"github.com/spf13/cobra"
)

type operationsCmd struct {
	*baseCmd
}

func newOperationsRootCmd() *operationsCmd {
	ccmd := &operationsCmd{}

	cmd := &cobra.Command{
		Use:   "operations",
		Short: "Cumulocity operations",
		Long:  `REST endpoint to interact with Cumulocity operations`,
	}

	// Subcommands
	cmd.AddCommand(newGetOperationCollectionCmd().getCommand())
	cmd.AddCommand(newGetOperationCmd().getCommand())
	cmd.AddCommand(newNewOperationCmd().getCommand())
	cmd.AddCommand(newUpdateOperationCmd().getCommand())
	cmd.AddCommand(newDeleteOperationCollectionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
