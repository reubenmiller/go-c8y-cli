package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type OperationsCmd struct {
	*subcommand.SubCommand
}

func NewOperationsRootCmd() *OperationsCmd {
	ccmd := &OperationsCmd{}

	cmd := &cobra.Command{
		Use:   "operations",
		Short: "Cumulocity operations",
		Long:  `REST endpoint to interact with Cumulocity operations`,
	}

	// Subcommands
	cmd.AddCommand(NewGetOperationCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetOperationCmd().GetCommand())
	cmd.AddCommand(NewNewOperationCmd().GetCommand())
	cmd.AddCommand(NewUpdateOperationCmd().GetCommand())
	cmd.AddCommand(NewDeleteOperationCollectionCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
