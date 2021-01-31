package cmd

import (
	"github.com/spf13/cobra"
)

type AuditRecordsCmd struct {
	*baseCmd
}

func NewAuditRecordsRootCmd() *AuditRecordsCmd {
	ccmd := &AuditRecordsCmd{}

	cmd := &cobra.Command{
		Use:   "auditRecords",
		Short: "Cumulocity auditRecords",
		Long:  `REST endpoint to interact with Cumulocity auditRecords`,
	}

	// Subcommands
	cmd.AddCommand(NewNewAuditCmd().getCommand())
	cmd.AddCommand(NewGetAuditRecordCollectionCmd().getCommand())
	cmd.AddCommand(NewGetAuditRecordCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
