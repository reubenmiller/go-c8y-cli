// Package databroker wires the `c8y databroker` command and its subcommands
// (list/get/update). All are hand-written v2 c8ystream commands, so this group
// command is itself hand-written rather than generated.
package databroker

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdDatabroker struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdDatabroker {
	ccmd := &SubCmdDatabroker{}

	cmd := &cobra.Command{
		Use:   "databroker",
		Short: "Cumulocity databroker",
		Long:  `REST endpoint to interact with Cumulocity data broker connectors`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
