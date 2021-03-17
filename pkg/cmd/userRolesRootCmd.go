package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type UserRolesCmd struct {
	*subcommand.SubCommand
}

func NewUserRolesRootCmd() *UserRolesCmd {
	ccmd := &UserRolesCmd{}

	cmd := &cobra.Command{
		Use:   "userRoles",
		Short: "Cumulocity user roles",
		Long:  `REST endpoint to interact with Cumulocity user roles`,
	}

	// Subcommands
	cmd.AddCommand(NewGetRoleCollectionCmd().GetCommand())
	cmd.AddCommand(NewAddRoleToUserCmd().GetCommand())
	cmd.AddCommand(NewDeleteRoleFromUserCmd().GetCommand())
	cmd.AddCommand(NewAddRoleToGroupCmd().GetCommand())
	cmd.AddCommand(NewDeleteRoleFromGroupCmd().GetCommand())
	cmd.AddCommand(NewGetRoleReferenceCollectionFromUserCmd().GetCommand())
	cmd.AddCommand(NewGetRoleReferenceCollectionFromGroupCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
