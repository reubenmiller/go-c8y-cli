// Package currentuser wires the `c8y currentuser` command and its subcommands
// (get/update/logout). All are hand-written v2 c8ystream commands, so this group
// command is itself hand-written rather than generated.
package currentuser

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdCurrentuser struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdCurrentuser {
	ccmd := &SubCmdCurrentuser{}

	cmd := &cobra.Command{
		Use:   "currentuser",
		Short: "Cumulocity current user",
		Long:  `REST endpoint to interact with the current Cumulocity user`,
	}

	// Subcommands
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewLogoutCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
