package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type AuditRecordsCmd struct {
	*subcommand.SubCommand
}

func NewAuditRecordsRootCmd() *AuditRecordsCmd {
	ccmd := &AuditRecordsCmd{}

	cmd := &cobra.Command{
		Use:   "auditRecords",
		Short: "Cumulocity auditRecords",
		Long:  `REST endpoint to interact with Cumulocity auditRecords`,
	}

	// Subcommands
	cmd.AddCommand(NewNewAuditCmd().GetCommand())
	cmd.AddCommand(NewGetAuditRecordCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetAuditRecordCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
