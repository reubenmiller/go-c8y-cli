package cmd

import (
	"github.com/spf13/cobra"
)

type OperationsCmd struct {
	*baseCmd
}

func NewOperationsRootCmd() *OperationsCmd {
	ccmd := &OperationsCmd{}

	cmd := &cobra.Command{
		Use:   "operations",
		Short: "Cumulocity operations",
		Long:  `REST endpoint to interact with Cumulocity operations`,
	}

	// Subcommands
	cmd.AddCommand(NewGetOperationCollectionCmd().getCommand())
	cmd.AddCommand(NewGetOperationCmd().getCommand())
	cmd.AddCommand(NewNewOperationCmd().getCommand())
	cmd.AddCommand(NewUpdateOperationCmd().getCommand())
	cmd.AddCommand(NewDeleteOperationCollectionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
