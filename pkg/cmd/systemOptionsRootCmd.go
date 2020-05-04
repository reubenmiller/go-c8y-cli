package cmd

import (
	"github.com/spf13/cobra"
)

type systemOptionsCmd struct {
	*baseCmd
}

func newSystemOptionsRootCmd() *systemOptionsCmd {
	ccmd := &systemOptionsCmd{}

	cmd := &cobra.Command{
		Use:   "systemOptions",
		Short: "Cumulocity systemOptions",
		Long:  `REST endpoint to interact with Cumulocity systemOptions`,
	}

	// Subcommands
	cmd.AddCommand(newGetSystemOptionCollectionCmd().getCommand())
	cmd.AddCommand(newGetSystemOptionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
