// Package tenantstatistics wires the `c8y tenantstatistics` command and its
// subcommands (list / listSummaryForTenant / listSummaryAllTenants). All are
// hand-written v2 c8ystream commands, so this group command is itself
// hand-written rather than generated.
package tenantstatistics

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/tenantstatistics/list"
	cmdListSummaryAllTenants "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/tenantstatistics/listsummaryalltenants"
	cmdListSummaryForTenant "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/tenantstatistics/listsummaryfortenant"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// SubCmdTenantstatistics is the `tenantstatistics` group command.
type SubCmdTenantstatistics struct {
	*subcommand.SubCommand
}

// NewSubCommand builds the `tenantstatistics` group command and attaches its
// subcommands.
func NewSubCommand(f *cmdutil.Factory) *SubCmdTenantstatistics {
	ccmd := &SubCmdTenantstatistics{}

	cmd := &cobra.Command{
		Use:   "tenantstatistics",
		Short: "Cumulocity tenant statistics",
		Long:  `REST endpoint to interact with Cumulocity tenant usage statistics`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdListSummaryAllTenants.NewListSummaryAllTenantsCmd(f).GetCommand())
	cmd.AddCommand(cmdListSummaryForTenant.NewListSummaryForTenantCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
