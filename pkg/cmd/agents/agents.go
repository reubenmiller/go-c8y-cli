// Package agents wires the `c8y agents` command and its subcommands. All
// subcommands are hand-written v2 c8ystream commands (backed by the go-c8y v2
// Agents service), so this group command is itself hand-written rather than
// generated.
package agents

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/agents/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/agents/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/agents/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/agents/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/agents/update"
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
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
