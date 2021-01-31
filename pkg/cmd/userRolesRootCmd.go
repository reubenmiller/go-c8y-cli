package cmd

import (
	"github.com/spf13/cobra"
)

type UserRolesCmd struct {
	*baseCmd
}

func NewUserRolesRootCmd() *UserRolesCmd {
	ccmd := &UserRolesCmd{}

	cmd := &cobra.Command{
		Use:   "userRoles",
		Short: "Cumulocity user roles",
		Long:  `REST endpoint to interact with Cumulocity user roles`,
	}

	// Subcommands
	cmd.AddCommand(NewGetRoleCollectionCmd().getCommand())
	cmd.AddCommand(NewAddRoleToUserCmd().getCommand())
	cmd.AddCommand(NewDeleteRoleFromUserCmd().getCommand())
	cmd.AddCommand(NewAddRoleToGroupCmd().getCommand())
	cmd.AddCommand(NewDeleteRoleFromGroupCmd().getCommand())
	cmd.AddCommand(NewGetRoleReferenceCollectionFromUserCmd().getCommand())
	cmd.AddCommand(NewGetRoleReferenceCollectionFromGroupCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
