// Package users wires the `c8y datahub users` command and its subcommands
// (list). All are hand-written v2 c8ystream commands, so this group command is
// itself hand-written rather than generated.
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
		Short: "Cumulocity DataHub users",
		Long:  `Cumulocity DataHub users`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
