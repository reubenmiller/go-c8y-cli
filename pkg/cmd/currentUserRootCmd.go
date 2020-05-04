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
	cmd.AddCommand(newGetUserCurrentCmd().getCommand())
	cmd.AddCommand(newUpdateUserCurrentCmd().getCommand())
	cmd.AddCommand(newGetUserCollectionCmd().getCommand())
	cmd.AddCommand(newNewUserCmd().getCommand())
	cmd.AddCommand(newGetUserCmd().getCommand())
	cmd.AddCommand(newGetUserByNameCmd().getCommand())
	cmd.AddCommand(newDeleteUserCmd().getCommand())
	cmd.AddCommand(newUpdateUserCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
