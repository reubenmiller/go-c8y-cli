package cmd

import (
	"github.com/spf13/cobra"
)

type agentsCmd struct {
	*baseCmd
}

func newAgentsRootCmd() *agentsCmd {
	ccmd := &agentsCmd{}

	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Cumulocity agents",
		Long:  `REST endpoint to interact with Cumulocity agents`,
	}

	// Subcommands
	cmd.AddCommand(newGetAgentCmd().getCommand())
	cmd.AddCommand(newUpdateAgentCmd().getCommand())
	cmd.AddCommand(newDeleteAgentCmd().getCommand())
	cmd.AddCommand(newCreateAgentCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
