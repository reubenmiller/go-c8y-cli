package cmd

import (
	"github.com/spf13/cobra"
)

type UsersCmd struct {
	*baseCmd
}

func NewUsersRootCmd() *UsersCmd {
	ccmd := &UsersCmd{}

	cmd := &cobra.Command{
		Use:   "users",
		Short: "Cumulocity users",
		Long:  `REST endpoint to interact with Cumulocity users`,
	}

	// Subcommands
	cmd.AddCommand(NewGetUserCurrentCmd().getCommand())
	cmd.AddCommand(NewUpdateUserCurrentCmd().getCommand())
	cmd.AddCommand(NewGetInventoryRoleCollectionCmd().getCommand())
	cmd.AddCommand(NewGetInventoryRoleCmd().getCommand())
	cmd.AddCommand(NewGetUserCollectionCmd().getCommand())
	cmd.AddCommand(NewNewUserCmd().getCommand())
	cmd.AddCommand(NewGetUserCmd().getCommand())
	cmd.AddCommand(NewGetUserByNameCmd().getCommand())
	cmd.AddCommand(NewDeleteUserCmd().getCommand())
	cmd.AddCommand(NewUpdateUserCmd().getCommand())
	cmd.AddCommand(NewResetUserPasswordCmd().getCommand())
	cmd.AddCommand(NewGetUserMembershipCollectionCmd().getCommand())
	cmd.AddCommand(NewLogoutCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
