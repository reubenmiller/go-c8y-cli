package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenantstatistics/list"
	cmdListSummaryAllTenants "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenantstatistics/listsummaryalltenants"
	cmdListSummaryForTenant "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenantstatistics/listsummaryfortenant"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdTenantstatistics struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdTenantstatistics {
	ccmd := &SubCmdTenantstatistics{}

	cmd := &cobra.Command{
		Use:   "tenantstatistics",
		Short: "Cumulocity tenant statistics",
		Long:  `REST endpoint to interact with Cumulocity tenant statistics`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdListSummaryAllTenants.NewListSummaryAllTenantsCmd(f).GetCommand())
	cmd.AddCommand(cmdListSummaryForTenant.NewListSummaryForTenantCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
