package cmd

import (
	"github.com/spf13/cobra"
)

type batchCmd struct {
	*baseCmd
}

func newBatchRootCmd() *batchCmd {
	ccmd := &batchCmd{}

	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Batch commands",
		Long:  `Create objects in parallel threads`,
	}
	cmd.Hidden = true

	// Subcommands
	cmd.AddCommand(newBatchCreateManagedObjectCmd().getCommand())
	cmd.AddCommand(newBatchAddDeviceToGroupCmd().getCommand())
	cmd.AddCommand(newBatchDeleteManagedObjectCmd().getCommand())
	cmd.AddCommand(newBatchCreateMeasurementCmd().getCommand())
	cmd.AddCommand(newBatchUpdateManagedObjectCmd().getCommand())
	cmd.AddCommand(newBatchDummyPipeCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
