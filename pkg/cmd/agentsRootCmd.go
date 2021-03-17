package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type AgentsCmd struct {
	*subcommand.SubCommand
}

func NewAgentsRootCmd() *AgentsCmd {
	ccmd := &AgentsCmd{}

	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Cumulocity agents",
		Long:  `REST endpoint to interact with Cumulocity agents`,
	}

	// Subcommands
	cmd.AddCommand(NewGetAgentCmd().GetCommand())
	cmd.AddCommand(NewUpdateAgentCmd().GetCommand())
	cmd.AddCommand(NewDeleteAgentCmd().GetCommand())
	cmd.AddCommand(NewCreateAgentCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
