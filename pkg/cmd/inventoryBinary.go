package cmd

import (
	"github.com/spf13/cobra"
)

type binaryManagedObjectCmd struct {
	*baseCmd
}

func newBinaryGetManagedObjectCmd() *binaryManagedObjectCmd {
	ccmd := &binaryManagedObjectCmd{}

	cmd := &cobra.Command{
		Use:   "binary",
		Short: "Inventory rest endpoint",
		Long:  `Inventory rest endpoint to interact with Cumulocity managed objects`,
	}

	cmd.AddCommand(newDownloadBinaryManagedObjectCmd().getCommand())
	cmd.AddCommand(newDeleteBinaryManagedObjectCmd().getCommand())
	cmd.AddCommand(newNewBinaryManagedObjectCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
