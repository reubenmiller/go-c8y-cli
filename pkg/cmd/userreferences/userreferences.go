// Package userreferences wires the `c8y userreferences` command group: user
// group membership — adding a user to a group, removing a user from a group, and
// listing the users in a group. Every subcommand is a hand-written v2 c8ystream
// command backed by the typed go-c8y v2 UserGroups.Users service.
package userreferences

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdUserreferences struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdUserreferences {
	ccmd := &SubCmdUserreferences{}

	cmd := &cobra.Command{
		Use:   "userreferences",
		Short: "Cumulocity user references",
		Long:  `REST endpoint to interact with Cumulocity user references`,
	}

	// Subcommands
	cmd.AddCommand(NewAddUserToGroupCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteUserFromGroupCmd(f).GetCommand())
	cmd.AddCommand(NewListGroupMembershipCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
