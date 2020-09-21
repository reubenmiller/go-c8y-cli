package cmd

import (
	"github.com/spf13/cobra"
)

type bulkOperationsCmd struct {
	*baseCmd
}

func newBulkOperationsRootCmd() *bulkOperationsCmd {
	ccmd := &bulkOperationsCmd{}

	cmd := &cobra.Command{
		Use:   "bulkOperations",
		Short: "Cumulocity bulk operations",
		Long:  `REST endpoint to interact with Cumulocity bulk operations`,
	}

	// Subcommands
	cmd.AddCommand(newGetBulkOperationCollectionCmd().getCommand())
	cmd.AddCommand(newGetBulkOperationCmd().getCommand())
	cmd.AddCommand(newNewBulkOperationCmd().getCommand())
	cmd.AddCommand(newUpdateBulkOperationCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
