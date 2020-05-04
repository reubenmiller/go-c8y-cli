package cmd

import (
	"github.com/spf13/cobra"
)

type userRolesCmd struct {
	*baseCmd
}

func newUserRolesRootCmd() *userRolesCmd {
	ccmd := &userRolesCmd{}

	cmd := &cobra.Command{
		Use:   "userRoles",
		Short: "Cumulocity user roles",
		Long:  `REST endpoint to interact with Cumulocity user roles`,
	}

	// Subcommands
	cmd.AddCommand(newGetRoleCollectionCmd().getCommand())
	cmd.AddCommand(newAddRoleToUserCmd().getCommand())
	cmd.AddCommand(newDeleteRoleFromUserCmd().getCommand())
	cmd.AddCommand(newAddRoleToGroupCmd().getCommand())
	cmd.AddCommand(newDeleteRoleFromGroupCmd().getCommand())
	cmd.AddCommand(newGetRoleReferenceCollectionFromUserCmd().getCommand())
	cmd.AddCommand(newGetRoleReferenceCollectionFromGroupCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
