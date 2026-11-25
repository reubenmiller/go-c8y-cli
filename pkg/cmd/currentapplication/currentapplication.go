// Package currentapplication wires the `c8y currentapplication` command and its
// subcommands (get/update/listSubscriptions). All are hand-written v2 c8ystream
// commands, so this group command is itself hand-written rather than generated.
package currentapplication

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdCurrentapplication struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdCurrentapplication {
	ccmd := &SubCmdCurrentapplication{}

	cmd := &cobra.Command{
		Use:   "currentapplication",
		Short: "Cumulocity currentApplication",
		Long:  `REST endpoint to interact with Cumulocity currentApplication`,
	}

	// Subcommands
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewListSubscriptionsCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
