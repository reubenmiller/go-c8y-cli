// Package userroles wires the `c8y userroles` command group: the role catalog
// (list) plus assigning, unassigning and listing role references on users and
// groups. Every subcommand is a hand-written v2 c8ystream command backed by the
// typed go-c8y v2 UserRoles service (UserRoles, UserRoles.Users, UserRoles.Groups).
package userroles

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdUserroles struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdUserroles {
	ccmd := &SubCmdUserroles{}

	cmd := &cobra.Command{
		Use:   "userroles",
		Short: "Cumulocity user roles",
		Long:  `REST endpoint to interact with Cumulocity user roles`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewAddRoleToUserCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteRoleFromUserCmd(f).GetCommand())
	cmd.AddCommand(NewAddRoleToGroupCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteRoleFromGroupCmd(f).GetCommand())
	cmd.AddCommand(NewGetRoleReferenceCollectionFromUserCmd(f).GetCommand())
	cmd.AddCommand(NewGetRoleReferenceCollectionFromGroupCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
