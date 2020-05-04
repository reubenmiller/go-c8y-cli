package cmd

import (
	"github.com/spf13/cobra"
)

type userReferencesCmd struct {
	*baseCmd
}

func newUserReferencesRootCmd() *userReferencesCmd {
	ccmd := &userReferencesCmd{}

	cmd := &cobra.Command{
		Use:   "userReferences",
		Short: "Cumulocity user references",
		Long:  `REST endpoint to interact with Cumulocity user references`,
	}

	// Subcommands
	cmd.AddCommand(newAddUserToGroupCmd().getCommand())
	cmd.AddCommand(newDeleteUserFromGroupCmd().getCommand())
	cmd.AddCommand(newGetUsersInGroupCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
