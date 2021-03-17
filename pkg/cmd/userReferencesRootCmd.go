package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type UserReferencesCmd struct {
	*subcommand.SubCommand
}

func NewUserReferencesRootCmd() *UserReferencesCmd {
	ccmd := &UserReferencesCmd{}

	cmd := &cobra.Command{
		Use:   "userReferences",
		Short: "Cumulocity user references",
		Long:  `REST endpoint to interact with Cumulocity user references`,
	}

	// Subcommands
	cmd.AddCommand(NewAddUserToGroupCmd().GetCommand())
	cmd.AddCommand(NewDeleteUserFromGroupCmd().GetCommand())
	cmd.AddCommand(NewGetUsersInGroupCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
