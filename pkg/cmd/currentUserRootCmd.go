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
	cmd.AddCommand(NewGetUserCollectionCmd().getCommand())
	cmd.AddCommand(NewNewUserCmd().getCommand())
	cmd.AddCommand(NewGetUserCmd().getCommand())
	cmd.AddCommand(NewGetUserByNameCmd().getCommand())
	cmd.AddCommand(NewDeleteUserCmd().getCommand())
	cmd.AddCommand(NewUpdateUserCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
