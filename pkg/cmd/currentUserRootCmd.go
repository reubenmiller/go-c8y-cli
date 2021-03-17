package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type currentUserCmd struct {
	*subcommand.SubCommand
}

func newCurrentUserRootCmd() *currentUserCmd {
	ccmd := &currentUserCmd{}

	cmd := &cobra.Command{
		Use:   "currentUser",
		Short: "Cumulocity current user",
		Long:  `REST endpoint to interact with Cumulocity current user`,
	}

	// Subcommands
	cmd.AddCommand(NewGetUserCurrentCmd().GetCommand())
	cmd.AddCommand(NewUpdateUserCurrentCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
