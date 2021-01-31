package cmd

import (
	"github.com/spf13/cobra"
)

type inventoryCmd struct {
	*baseCmd
}

func newInventoryRootCmd() *inventoryCmd {
	ccmd := &inventoryCmd{}

	cmd := &cobra.Command{
		Use:   "inventory",
		Short: "Cumulocity managed objects",
		Long:  `REST endpoint to interact with Cumulocity managed objects`,
	}

	// Subcommands
	cmd.AddCommand(newGetManagedObjectCollectionCmd().getCommand())
	cmd.AddCommand(newNewManagedObjectCmd().getCommand())
	cmd.AddCommand(newGetManagedObjectCmd().getCommand())
	cmd.AddCommand(newUpdateManagedObjectCmd().getCommand())
	cmd.AddCommand(newDeleteManagedObjectCmd().getCommand())
	cmd.AddCommand(newGetManagedObjectCmdManual().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
