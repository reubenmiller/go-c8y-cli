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
	cmd.AddCommand(NewGetGroupCollectionCmd().getCommand())
	cmd.AddCommand(NewNewGroupCmd().getCommand())
	cmd.AddCommand(NewGetGroupCmd().getCommand())
	cmd.AddCommand(NewGetGroupByNameCmd().getCommand())
	cmd.AddCommand(NewDeleteGroupCmd().getCommand())
	cmd.AddCommand(NewUpdateGroupCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
