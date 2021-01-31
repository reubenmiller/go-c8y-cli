package cmd

import (
	"github.com/spf13/cobra"
)

type UserReferencesCmd struct {
	*baseCmd
}

func NewUserReferencesRootCmd() *UserReferencesCmd {
	ccmd := &UserReferencesCmd{}

	cmd := &cobra.Command{
		Use:   "userReferences",
		Short: "Cumulocity user references",
		Long:  `REST endpoint to interact with Cumulocity user references`,
	}

	// Subcommands
	cmd.AddCommand(NewAddUserToGroupCmd().getCommand())
	cmd.AddCommand(NewDeleteUserFromGroupCmd().getCommand())
	cmd.AddCommand(NewGetUsersInGroupCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
