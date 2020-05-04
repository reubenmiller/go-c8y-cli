package cmd

import (
	"github.com/spf13/cobra"
)

type binariesCmd struct {
	*baseCmd
}

func newBinariesRootCmd() *binariesCmd {
	ccmd := &binariesCmd{}

	cmd := &cobra.Command{
		Use:   "binaries",
		Short: "Cumulocity binaries",
		Long:  `REST endpoint to interact with Cumulocity binaries`,
	}

	// Subcommands
	cmd.AddCommand(newGetBinaryCollectionCmd().getCommand())
	cmd.AddCommand(newDownloadCmd().getCommand())
	cmd.AddCommand(newNewBinaryCmd().getCommand())
	cmd.AddCommand(newUpdateBinaryCmd().getCommand())
	cmd.AddCommand(newDeleteBinaryCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
