// Package systemoptions wires the `c8y systemoptions` command and its
// subcommands (list/get). Both are hand-written v2 c8ystream commands, so this
// group command is itself hand-written rather than generated.
package systemoptions

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/systemoptions/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/systemoptions/list"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// SubCmdSystemoptions is the `systemoptions` group command.
type SubCmdSystemoptions struct {
	*subcommand.SubCommand
}

// NewSubCommand builds the `systemoptions` group command and attaches its
// subcommands.
func NewSubCommand(f *cmdutil.Factory) *SubCmdSystemoptions {
	ccmd := &SubCmdSystemoptions{}

	cmd := &cobra.Command{
		Use:   "systemoptions",
		Short: "Cumulocity system options",
		Long:  `REST endpoint to interact with Cumulocity system options`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
