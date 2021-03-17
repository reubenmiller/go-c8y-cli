package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type SystemOptionsCmd struct {
	*subcommand.SubCommand
}

func NewSystemOptionsRootCmd() *SystemOptionsCmd {
	ccmd := &SystemOptionsCmd{}

	cmd := &cobra.Command{
		Use:   "systemOptions",
		Short: "Cumulocity systemOptions",
		Long:  `REST endpoint to interact with Cumulocity systemOptions`,
	}

	// Subcommands
	cmd.AddCommand(NewGetSystemOptionCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetSystemOptionCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
