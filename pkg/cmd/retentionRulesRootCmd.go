package cmd

import (
	"github.com/spf13/cobra"
)

type retentionRulesCmd struct {
	*baseCmd
}

func newRetentionRulesRootCmd() *retentionRulesCmd {
	ccmd := &retentionRulesCmd{}

	cmd := &cobra.Command{
		Use:   "retentionRules",
		Short: "Cumulocity retentionRules",
		Long:  `REST endpoint to interact with Cumulocity retentionRules`,
	}

	// Subcommands
	cmd.AddCommand(newGetRetentionRuleCollectionCmd().getCommand())
	cmd.AddCommand(newNewRetentionRuleCmd().getCommand())
	cmd.AddCommand(newGetRetentionRuleCmd().getCommand())
	cmd.AddCommand(newDeleteRetentionRuleCmd().getCommand())
	cmd.AddCommand(newUpdateRetentionRuleCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
