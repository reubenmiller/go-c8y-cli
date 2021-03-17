package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type InventoryCmd struct {
	*subcommand.SubCommand
}

func NewInventoryRootCmd() *InventoryCmd {
	ccmd := &InventoryCmd{}

	cmd := &cobra.Command{
		Use:   "inventory",
		Short: "Cumulocity managed objects",
		Long:  `REST endpoint to interact with Cumulocity managed objects`,
	}

	// Subcommands
	cmd.AddCommand(NewGetManagedObjectCollectionCmd().GetCommand())
	cmd.AddCommand(NewNewManagedObjectCmd().GetCommand())
	cmd.AddCommand(NewGetManagedObjectCmd().GetCommand())
	cmd.AddCommand(NewUpdateManagedObjectCmd().GetCommand())
	cmd.AddCommand(NewDeleteManagedObjectCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
