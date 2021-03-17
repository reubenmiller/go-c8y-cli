package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type UsersCmd struct {
	*subcommand.SubCommand
}

func NewUsersRootCmd() *UsersCmd {
	ccmd := &UsersCmd{}

	cmd := &cobra.Command{
		Use:   "users",
		Short: "Cumulocity users",
		Long:  `REST endpoint to interact with Cumulocity users`,
	}

	// Subcommands
	cmd.AddCommand(NewGetUserCurrentCmd().GetCommand())
	cmd.AddCommand(NewUpdateUserCurrentCmd().GetCommand())
	cmd.AddCommand(NewGetInventoryRoleCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetInventoryRoleCmd().GetCommand())
	cmd.AddCommand(NewGetUserCollectionCmd().GetCommand())
	cmd.AddCommand(NewNewUserCmd().GetCommand())
	cmd.AddCommand(NewGetUserCmd().GetCommand())
	cmd.AddCommand(NewGetUserByNameCmd().GetCommand())
	cmd.AddCommand(NewDeleteUserCmd().GetCommand())
	cmd.AddCommand(NewUpdateUserCmd().GetCommand())
	cmd.AddCommand(NewResetUserPasswordCmd().GetCommand())
	cmd.AddCommand(NewGetUserMembershipCollectionCmd().GetCommand())
	cmd.AddCommand(NewLogoutCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
