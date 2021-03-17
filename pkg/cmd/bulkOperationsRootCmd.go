package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type BulkOperationsCmd struct {
	*subcommand.SubCommand
}

func NewBulkOperationsRootCmd() *BulkOperationsCmd {
	ccmd := &BulkOperationsCmd{}

	cmd := &cobra.Command{
		Use:   "bulkOperations",
		Short: "Cumulocity bulk operations",
		Long:  `REST endpoint to interact with Cumulocity bulk operations`,
	}

	// Subcommands
	cmd.AddCommand(NewGetBulkOperationCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetBulkOperationCmd().GetCommand())
	cmd.AddCommand(NewDeleteBulkOperationCmd().GetCommand())
	cmd.AddCommand(NewNewBulkOperationCmd().GetCommand())
	cmd.AddCommand(NewUpdateBulkOperationCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
