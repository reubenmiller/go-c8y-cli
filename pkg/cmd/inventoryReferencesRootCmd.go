package cmd

import (
	"github.com/spf13/cobra"
)

type InventoryReferencesCmd struct {
	*baseCmd
}

func NewInventoryReferencesRootCmd() *InventoryReferencesCmd {
	ccmd := &InventoryReferencesCmd{}

	cmd := &cobra.Command{
		Use:   "inventoryReferences",
		Short: "Cumulocity managed objects",
		Long:  `REST endpoint to interact with Cumulocity managed objects`,
	}

	// Subcommands
	cmd.AddCommand(NewGetManagedObjectChildDeviceCollectionCmd().getCommand())
	cmd.AddCommand(NewGetManagedObjectChildAssetCollectionCmd().getCommand())
	cmd.AddCommand(NewNewManagedObjectChildDeviceCmd().getCommand())
	cmd.AddCommand(NewAddDeviceToGroupCmd().getCommand())
	cmd.AddCommand(NewAddGroupToGroupCmd().getCommand())
	cmd.AddCommand(NewNewManagedObjectChildAssetCmd().getCommand())
	cmd.AddCommand(NewGetManagedObjectChildDeviceReferenceCmd().getCommand())
	cmd.AddCommand(NewGetManagedObjectChildAssetReferenceCmd().getCommand())
	cmd.AddCommand(NewDeleteManagedObjectChildDeviceReferenceCmd().getCommand())
	cmd.AddCommand(NewDeleteManagedObjectChildAssetReferenceCmd().getCommand())
	cmd.AddCommand(NewDeleteDeviceFromGroupCmd().getCommand())
	cmd.AddCommand(NewDeleteAssetFromGroupCmd().getCommand())
	cmd.AddCommand(NewGetManagedObjectChildAdditionCollectionCmd().getCommand())
	cmd.AddCommand(NewAddManagedObjectChildAdditionCmd().getCommand())
	cmd.AddCommand(NewDeleteChildAdditionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
