package cmd

import (
	"github.com/spf13/cobra"
)

type BinariesCmd struct {
	*baseCmd
}

func NewBinariesRootCmd() *BinariesCmd {
	ccmd := &BinariesCmd{}

	cmd := &cobra.Command{
		Use:   "binaries",
		Short: "Cumulocity binaries",
		Long:  `REST endpoint to interact with Cumulocity binaries`,
	}

	// Subcommands
	cmd.AddCommand(NewGetBinaryCollectionCmd().getCommand())
	cmd.AddCommand(NewDownloadCmd().getCommand())
	cmd.AddCommand(NewNewBinaryCmd().getCommand())
	cmd.AddCommand(NewUpdateBinaryCmd().getCommand())
	cmd.AddCommand(NewDeleteBinaryCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
