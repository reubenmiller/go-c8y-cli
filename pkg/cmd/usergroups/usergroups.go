// Package usergroups wires the `c8y usergroups` command and its subcommands.
// All subcommands are hand-written v2 c8ystream commands (backed by the go-c8y v2
// UserGroups service), so this group command is itself hand-written rather than
// generated.
package usergroups

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdUsergroups struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdUsergroups {
	ccmd := &SubCmdUsergroups{}

	cmd := &cobra.Command{
		Use:   "usergroups",
		Short: "Cumulocity user groups",
		Long:  `REST endpoint to interact with Cumulocity user groups`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewGetByNameCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
