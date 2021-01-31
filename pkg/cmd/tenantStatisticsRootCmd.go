package cmd

import (
	"github.com/spf13/cobra"
)

type TenantStatisticsCmd struct {
	*baseCmd
}

func NewTenantStatisticsRootCmd() *TenantStatisticsCmd {
	ccmd := &TenantStatisticsCmd{}

	cmd := &cobra.Command{
		Use:   "tenantStatistics",
		Short: "Cumulocity tenant statistics",
		Long:  `REST endpoint to interact with Cumulocity tenant statistics`,
	}

	// Subcommands
	cmd.AddCommand(NewGetTenantUsageStatisticsCollectionCmd().getCommand())
	cmd.AddCommand(NewGetAllTenantUsageStatisticsSummaryCollectionCmd().getCommand())
	cmd.AddCommand(NewGetTenantUsageStatisticsSummaryCollectionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
