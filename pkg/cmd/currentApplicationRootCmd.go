package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type CurrentApplicationCmd struct {
	*subcommand.SubCommand
}

func NewCurrentApplicationRootCmd() *CurrentApplicationCmd {
	ccmd := &CurrentApplicationCmd{}

	cmd := &cobra.Command{
		Use:   "currentApplication",
		Short: "Cumulocity currentApplication",
		Long:  `REST endpoint to interact with Cumulocity currentApplication`,
	}

	// Subcommands
	cmd.AddCommand(NewGetCurrentApplicationCmd().GetCommand())
	cmd.AddCommand(NewUpdateCurrentApplicationCmd().GetCommand())
	cmd.AddCommand(NewGetCurrentApplicationUserCollectionCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
