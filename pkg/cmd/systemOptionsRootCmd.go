package cmd

import (
	"github.com/spf13/cobra"
)

type SystemOptionsCmd struct {
	*baseCmd
}

func NewSystemOptionsRootCmd() *SystemOptionsCmd {
	ccmd := &SystemOptionsCmd{}

	cmd := &cobra.Command{
		Use:   "systemOptions",
		Short: "Cumulocity systemOptions",
		Long:  `REST endpoint to interact with Cumulocity systemOptions`,
	}

	// Subcommands
	cmd.AddCommand(NewGetSystemOptionCollectionCmd().getCommand())
	cmd.AddCommand(NewGetSystemOptionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
