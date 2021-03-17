package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type RetentionRulesCmd struct {
	*subcommand.SubCommand
}

func NewRetentionRulesRootCmd() *RetentionRulesCmd {
	ccmd := &RetentionRulesCmd{}

	cmd := &cobra.Command{
		Use:   "retentionRules",
		Short: "Cumulocity retentionRules",
		Long:  `REST endpoint to interact with Cumulocity retentionRules`,
	}

	// Subcommands
	cmd.AddCommand(NewGetRetentionRuleCollectionCmd().GetCommand())
	cmd.AddCommand(NewNewRetentionRuleCmd().GetCommand())
	cmd.AddCommand(NewGetRetentionRuleCmd().GetCommand())
	cmd.AddCommand(NewDeleteRetentionRuleCmd().GetCommand())
	cmd.AddCommand(NewUpdateRetentionRuleCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
