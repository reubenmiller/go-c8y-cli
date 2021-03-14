package cmd

import (
	"github.com/spf13/cobra"
)

type currentUserCmd struct {
	*baseCmd
}

func newCurrentUserRootCmd() *currentUserCmd {
	ccmd := &currentUserCmd{}

	cmd := &cobra.Command{
		Use:   "currentUser",
		Short: "Cumulocity current user",
		Long:  `REST endpoint to interact with Cumulocity current user`,
	}

	// Subcommands
	cmd.AddCommand(NewGetUserCurrentCmd().getCommand())
	cmd.AddCommand(NewUpdateUserCurrentCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
