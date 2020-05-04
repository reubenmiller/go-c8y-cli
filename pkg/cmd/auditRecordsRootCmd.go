package cmd

import (
	"github.com/spf13/cobra"
)

type auditRecordsCmd struct {
	*baseCmd
}

func newAuditRecordsRootCmd() *auditRecordsCmd {
	ccmd := &auditRecordsCmd{}

	cmd := &cobra.Command{
		Use:   "auditRecords",
		Short: "Cumulocity auditRecords",
		Long:  `REST endpoint to interact with Cumulocity auditRecords`,
	}

	// Subcommands
	cmd.AddCommand(newNewAuditCmd().getCommand())
	cmd.AddCommand(newGetAuditRecordCollectionCmd().getCommand())
	cmd.AddCommand(newGetAuditRecordCmd().getCommand())
	cmd.AddCommand(newDeleteAuditRecordCollectionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
