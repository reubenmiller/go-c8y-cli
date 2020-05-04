package cmd

import (
	"github.com/spf13/cobra"
)

type tenantStatisticsCmd struct {
	*baseCmd
}

func newTenantStatisticsRootCmd() *tenantStatisticsCmd {
	ccmd := &tenantStatisticsCmd{}

	cmd := &cobra.Command{
		Use:   "tenantStatistics",
		Short: "Cumulocity tenant statistics",
		Long:  `REST endpoint to interact with Cumulocity tenant statistics`,
	}

	// Subcommands
	cmd.AddCommand(newGetTenantUsageStatisticsCollectionCmd().getCommand())
	cmd.AddCommand(newGetAllTenantUsageStatisticsSummaryCollectionCmd().getCommand())
	cmd.AddCommand(newGetTenantUsageStatisticsSummaryCollectionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
