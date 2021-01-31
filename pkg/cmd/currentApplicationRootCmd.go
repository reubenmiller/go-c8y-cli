package cmd

import (
	"github.com/spf13/cobra"
)

type CurrentApplicationCmd struct {
	*baseCmd
}

func NewCurrentApplicationRootCmd() *CurrentApplicationCmd {
	ccmd := &CurrentApplicationCmd{}

	cmd := &cobra.Command{
		Use:   "currentApplication",
		Short: "Cumulocity currentApplication",
		Long:  `REST endpoint to interact with Cumulocity currentApplication`,
	}

	// Subcommands
	cmd.AddCommand(NewGetCurrentApplicationCmd().getCommand())
	cmd.AddCommand(NewUpdateCurrentApplicationCmd().getCommand())
	cmd.AddCommand(NewGetCurrentApplicationUserCollectionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
