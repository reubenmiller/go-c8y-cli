package cmd

import (
	"github.com/spf13/cobra"
)

type UserGroupsCmd struct {
	*baseCmd
}

func NewUserGroupsRootCmd() *UserGroupsCmd {
	ccmd := &UserGroupsCmd{}

	cmd := &cobra.Command{
		Use:   "userGroups",
		Short: "Cumulocity user groups",
		Long:  `REST endpoint to interact with Cumulocity user groups`,
	}

	// Subcommands
	cmd.AddCommand(NewGetUserGroupCollectionCmd().getCommand())
	cmd.AddCommand(NewCreateUserGroupCmd().getCommand())
	cmd.AddCommand(NewGetUserGroupCmd().getCommand())
	cmd.AddCommand(NewGetUserGroupByNameCmd().getCommand())
	cmd.AddCommand(NewDeleteUserGroupCmd().getCommand())
	cmd.AddCommand(NewUpdateUserGroupCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
