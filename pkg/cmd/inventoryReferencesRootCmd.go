package cmd

import (
	"github.com/spf13/cobra"
)

type inventoryReferencesCmd struct {
	*baseCmd
}

func newInventoryReferencesRootCmd() *inventoryReferencesCmd {
	ccmd := &inventoryReferencesCmd{}

	cmd := &cobra.Command{
		Use:   "inventoryReferences",
		Short: "Cumulocity managed objects",
		Long:  `REST endpoint to interact with Cumulocity managed objects`,
	}

	// Subcommands
	cmd.AddCommand(newGetManagedObjectChildDeviceCollectionCmd().getCommand())
	cmd.AddCommand(newGetManagedObjectChildAssetCollectionCmd().getCommand())
	cmd.AddCommand(newNewManagedObjectChildDeviceCmd().getCommand())
	cmd.AddCommand(newAddDeviceToGroupCmd().getCommand())
	cmd.AddCommand(newAddGroupToGroupCmd().getCommand())
	cmd.AddCommand(newNewManagedObjectChildAssetCmd().getCommand())
	cmd.AddCommand(newGetManagedObjectChildDeviceReferenceCmd().getCommand())
	cmd.AddCommand(newGetManagedObjectChildAssetReferenceCmd().getCommand())
	cmd.AddCommand(newDeleteManagedObjectChildDeviceReferenceCmd().getCommand())
	cmd.AddCommand(newDeleteManagedObjectChildAssetReferenceCmd().getCommand())
	cmd.AddCommand(newDeleteDeviceFromGroupCmd().getCommand())
	cmd.AddCommand(newDeleteAssetFromGroupCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
