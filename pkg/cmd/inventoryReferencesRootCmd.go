package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type InventoryReferencesCmd struct {
	*subcommand.SubCommand
}

func NewInventoryReferencesRootCmd() *InventoryReferencesCmd {
	ccmd := &InventoryReferencesCmd{}

	cmd := &cobra.Command{
		Use:   "inventoryReferences",
		Short: "Cumulocity inventory references",
		Long:  `REST endpoint to interact with Cumulocity managed objects`,
	}

	// Subcommands
	cmd.AddCommand(NewGetManagedObjectChildDeviceCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetManagedObjectChildAssetCollectionCmd().GetCommand())
	cmd.AddCommand(NewNewManagedObjectChildDeviceCmd().GetCommand())
	cmd.AddCommand(NewAddDeviceToGroupCmd().GetCommand())
	cmd.AddCommand(NewAddGroupToGroupCmd().GetCommand())
	cmd.AddCommand(NewNewManagedObjectChildAssetCmd().GetCommand())
	cmd.AddCommand(NewGetManagedObjectChildDeviceReferenceCmd().GetCommand())
	cmd.AddCommand(NewGetManagedObjectChildAssetReferenceCmd().GetCommand())
	cmd.AddCommand(NewDeleteManagedObjectChildDeviceReferenceCmd().GetCommand())
	cmd.AddCommand(NewDeleteManagedObjectChildAssetReferenceCmd().GetCommand())
	cmd.AddCommand(NewDeleteDeviceFromGroupCmd().GetCommand())
	cmd.AddCommand(NewDeleteAssetFromGroupCmd().GetCommand())
	cmd.AddCommand(NewGetManagedObjectChildAdditionCollectionCmd().GetCommand())
	cmd.AddCommand(NewAddManagedObjectChildAdditionCmd().GetCommand())
	cmd.AddCommand(NewDeleteChildAdditionCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
