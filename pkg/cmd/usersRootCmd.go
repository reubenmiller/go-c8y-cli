package cmd

import (
	"github.com/spf13/cobra"
)

type usersCmd struct {
	*baseCmd
}

func newUsersRootCmd() *usersCmd {
	ccmd := &usersCmd{}

	cmd := &cobra.Command{
		Use:   "users",
		Short: "Cumulocity users",
		Long:  `REST endpoint to interact with Cumulocity users`,
	}

	// Subcommands
	cmd.AddCommand(newGetUserCurrentCmd().getCommand())
	cmd.AddCommand(newUpdateUserCurrentCmd().getCommand())
	cmd.AddCommand(newGetCurrentUserInventoryRoleCollectionCmd().getCommand())
	cmd.AddCommand(newGetCurrentUserInventoryRoleCmd().getCommand())
	cmd.AddCommand(newGetUserCollectionCmd().getCommand())
	cmd.AddCommand(newNewUserCmd().getCommand())
	cmd.AddCommand(newGetUserCmd().getCommand())
	cmd.AddCommand(newGetUserByNameCmd().getCommand())
	cmd.AddCommand(newDeleteUserCmd().getCommand())
	cmd.AddCommand(newUpdateUserCmd().getCommand())
	cmd.AddCommand(newResetUserPasswordCmd().getCommand())
	cmd.AddCommand(newGetUserMembershipCollectionCmd().getCommand())
	cmd.AddCommand(newLogoutCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
