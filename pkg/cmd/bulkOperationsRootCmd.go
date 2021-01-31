package cmd

import (
	"github.com/spf13/cobra"
)

type BulkOperationsCmd struct {
	*baseCmd
}

func NewBulkOperationsRootCmd() *BulkOperationsCmd {
	ccmd := &BulkOperationsCmd{}

	cmd := &cobra.Command{
		Use:   "bulkOperations",
		Short: "Cumulocity bulk operations",
		Long:  `REST endpoint to interact with Cumulocity bulk operations`,
	}

	// Subcommands
	cmd.AddCommand(NewGetBulkOperationCollectionCmd().getCommand())
	cmd.AddCommand(NewGetBulkOperationCmd().getCommand())
	cmd.AddCommand(NewDeleteBulkOperationCmd().getCommand())
	cmd.AddCommand(NewNewBulkOperationCmd().getCommand())
	cmd.AddCommand(NewUpdateBulkOperationCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
