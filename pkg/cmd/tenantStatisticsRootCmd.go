package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type TenantStatisticsCmd struct {
	*subcommand.SubCommand
}

func NewTenantStatisticsRootCmd() *TenantStatisticsCmd {
	ccmd := &TenantStatisticsCmd{}

	cmd := &cobra.Command{
		Use:   "tenantStatistics",
		Short: "Cumulocity tenant statistics",
		Long:  `REST endpoint to interact with Cumulocity tenant statistics`,
	}

	// Subcommands
	cmd.AddCommand(NewGetTenantUsageStatisticsCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetAllTenantUsageStatisticsSummaryCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetTenantUsageStatisticsSummaryCollectionCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
