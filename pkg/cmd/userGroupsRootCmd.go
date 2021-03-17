package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type UserGroupsCmd struct {
	*subcommand.SubCommand
}

func NewUserGroupsRootCmd() *UserGroupsCmd {
	ccmd := &UserGroupsCmd{}

	cmd := &cobra.Command{
		Use:   "userGroups",
		Short: "Cumulocity user groups",
		Long:  `REST endpoint to interact with Cumulocity user groups`,
	}

	// Subcommands
	cmd.AddCommand(NewGetUserGroupCollectionCmd().GetCommand())
	cmd.AddCommand(NewCreateUserGroupCmd().GetCommand())
	cmd.AddCommand(NewGetUserGroupCmd().GetCommand())
	cmd.AddCommand(NewGetUserGroupByNameCmd().GetCommand())
	cmd.AddCommand(NewDeleteUserGroupCmd().GetCommand())
	cmd.AddCommand(NewUpdateUserGroupCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
