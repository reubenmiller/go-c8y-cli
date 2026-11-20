// Package users wires the `c8y users` command and its subcommands.
// All subcommands are hand-written v2 c8ystream commands (backed by the go-c8y v2
// Users and InventoryRoles services), so this group command is itself
// hand-written rather than generated.
package users

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdUsers struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdUsers {
	ccmd := &SubCmdUsers{}

	cmd := &cobra.Command{
		Use:   "users",
		Short: "Cumulocity users",
		Long:  `REST endpoint to interact with Cumulocity users`,
	}

	// Subcommands
	cmd.AddCommand(NewListInventoryRolesCmd(f).GetCommand())
	cmd.AddCommand(NewGetInventoryRoleCmd(f).GetCommand())
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewGetUserByNameCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewResetUserPasswordCmd(f).GetCommand())
	cmd.AddCommand(NewRevokeTOTPSecretCmd(f).GetCommand())
	cmd.AddCommand(NewListUserMembershipCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
