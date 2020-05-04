package cmd

import (
	"github.com/spf13/cobra"
)

type currentApplicationCmd struct {
	*baseCmd
}

func newCurrentApplicationRootCmd() *currentApplicationCmd {
	ccmd := &currentApplicationCmd{}

	cmd := &cobra.Command{
		Use:   "currentApplication",
		Short: "Cumulocity currentApplication",
		Long:  `REST endpoint to interact with Cumulocity currentApplication`,
	}

	// Subcommands
	cmd.AddCommand(newGetCurrentApplicationCmd().getCommand())
	cmd.AddCommand(newUpdateCurrentApplicationCmd().getCommand())
	cmd.AddCommand(newGetCurrentApplicationUserCollectionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
