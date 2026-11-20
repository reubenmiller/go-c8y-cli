// Package agents wires the `c8y agents` command and its subcommands. All
// subcommands are hand-written v2 c8ystream commands (backed by the go-c8y v2
// Agents service), so this group command is itself hand-written rather than
// generated.
package agents

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAgents struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdAgents {
	ccmd := &SubCmdAgents{}

	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Cumulocity agents",
		Long:  `REST endpoint to interact with Cumulocity agents`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
