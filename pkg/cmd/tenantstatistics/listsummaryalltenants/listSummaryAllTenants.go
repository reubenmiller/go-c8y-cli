// v2-based tenant statistics summary (all tenants): returns the usage summary
// across all tenants for the given date window via
// Tenants.UsageStatistics.ListSummaryAllTenants. Runs once.
package listsummaryalltenants

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/usagestatistics"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// ListSummaryAllTenantsCmd command
type ListSummaryAllTenantsCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListSummaryAllTenantsCmd creates a command to Get all tenant usage summary statistics
func NewListSummaryAllTenantsCmd(f *cmdutil.Factory) *ListSummaryAllTenantsCmd {
	ccmd := &ListSummaryAllTenantsCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "listSummaryAllTenants",
		Short: "Get all tenant usage summary statistics",
		Long:  `Get the usage summary statistics of all tenants`,
		Example: heredoc.Doc(`
$ c8y tenantstatistics listSummaryAllTenants
Get tenant summary statistics for all tenants

$ c8y tenantstatistics listSummaryAllTenants --dateFrom "-30d"
Get tenant summary statistics for all tenants for the last 30 days
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("dateFrom", "", "Start date or date and time of the statistics.")
	cmd.Flags().String("dateTo", "", "End date or date and time of the statistics.")

	flags.WithOptions(
		cmd,
		flags.WithPowershellName("Get-AllTenantUsageSummaryStatistics"),
		flags.WithOutputType("application/json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListSummaryAllTenantsCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := usagestatistics.ListSummaryAllTenantsOptions{
			DateFrom: in.TimeValue("dateFrom"),
			DateTo:   in.TimeValue("dateTo"),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.Tenants.UsageStatistics.ListSummaryAllTenants(ctx, opt))
		}, nil
	})
}
