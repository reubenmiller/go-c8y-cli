package cmd

import (
	"github.com/spf13/cobra"
)

type AgentsCmd struct {
	*baseCmd
}

func NewAgentsRootCmd() *AgentsCmd {
	ccmd := &AgentsCmd{}

	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Cumulocity agents",
		Long:  `REST endpoint to interact with Cumulocity agents`,
	}

	// Subcommands
	cmd.AddCommand(NewGetAgentCmd().getCommand())
	cmd.AddCommand(NewUpdateAgentCmd().getCommand())
	cmd.AddCommand(NewDeleteAgentCmd().getCommand())
	cmd.AddCommand(NewCreateAgentCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
