package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type BinariesCmd struct {
	*subcommand.SubCommand
}

func NewBinariesRootCmd() *BinariesCmd {
	ccmd := &BinariesCmd{}

	cmd := &cobra.Command{
		Use:   "binaries",
		Short: "Cumulocity binaries",
		Long:  `REST endpoint to interact with Cumulocity binaries`,
	}

	// Subcommands
	cmd.AddCommand(NewGetBinaryCollectionCmd().GetCommand())
	cmd.AddCommand(NewDownloadCmd().GetCommand())
	cmd.AddCommand(NewNewBinaryCmd().GetCommand())
	cmd.AddCommand(NewUpdateBinaryCmd().GetCommand())
	cmd.AddCommand(NewDeleteBinaryCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
