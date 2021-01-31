package cmd

import (
	"github.com/spf13/cobra"
)

type InventoryCmd struct {
	*baseCmd
}

func NewInventoryRootCmd() *InventoryCmd {
	ccmd := &InventoryCmd{}

	cmd := &cobra.Command{
		Use:   "inventory",
		Short: "Cumulocity managed objects",
		Long:  `REST endpoint to interact with Cumulocity managed objects`,
	}

	// Subcommands
	cmd.AddCommand(NewGetManagedObjectCollectionCmd().getCommand())
	cmd.AddCommand(NewNewManagedObjectCmd().getCommand())
	cmd.AddCommand(NewGetManagedObjectCmd().getCommand())
	cmd.AddCommand(NewUpdateManagedObjectCmd().getCommand())
	cmd.AddCommand(NewDeleteManagedObjectCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
