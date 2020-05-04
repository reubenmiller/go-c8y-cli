package cmd

import (
	"github.com/spf13/cobra"
)

type userGroupsCmd struct {
	*baseCmd
}

func newUserGroupsRootCmd() *userGroupsCmd {
	ccmd := &userGroupsCmd{}

	cmd := &cobra.Command{
		Use:   "userGroups",
		Short: "Cumulocity user groups",
		Long:  `REST endpoint to interact with Cumulocity user groups`,
	}

	// Subcommands
	cmd.AddCommand(newGetGroupCollectionCmd().getCommand())
	cmd.AddCommand(newNewGroupCmd().getCommand())
	cmd.AddCommand(newGetGroupCmd().getCommand())
	cmd.AddCommand(newGetGroupByNameCmd().getCommand())
	cmd.AddCommand(newDeleteGroupCmd().getCommand())
	cmd.AddCommand(newUpdateGroupCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
