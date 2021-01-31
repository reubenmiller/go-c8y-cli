package cmd

import (
	"github.com/spf13/cobra"
)

type RetentionRulesCmd struct {
	*baseCmd
}

func NewRetentionRulesRootCmd() *RetentionRulesCmd {
	ccmd := &RetentionRulesCmd{}

	cmd := &cobra.Command{
		Use:   "retentionRules",
		Short: "Cumulocity retentionRules",
		Long:  `REST endpoint to interact with Cumulocity retentionRules`,
	}

	// Subcommands
	cmd.AddCommand(NewGetRetentionRuleCollectionCmd().getCommand())
	cmd.AddCommand(NewNewRetentionRuleCmd().getCommand())
	cmd.AddCommand(NewGetRetentionRuleCmd().getCommand())
	cmd.AddCommand(NewDeleteRetentionRuleCmd().getCommand())
	cmd.AddCommand(NewUpdateRetentionRuleCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
