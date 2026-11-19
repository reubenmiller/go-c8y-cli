// v2-based tenant statistics summary (current tenant): returns the usage summary
// for the given date window via Tenants.UsageStatistics.ListSummary. Runs once.
package listsummaryfortenant

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

// ListSummaryForTenantCmd command
type ListSummaryForTenantCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListSummaryForTenantCmd creates a command to Get tenant usage summary statistics
func NewListSummaryForTenantCmd(f *cmdutil.Factory) *ListSummaryForTenantCmd {
	ccmd := &ListSummaryForTenantCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "listSummaryForTenant",
		Short: "Get tenant usage summary statistics",
		Long:  `Get the usage summary statistics of the current tenant`,
		Example: heredoc.Doc(`
$ c8y tenantstatistics listSummaryForTenant
Get tenant summary statistics for the current tenant

$ c8y tenantstatistics listSummaryForTenant --dateFrom "-30d"
Get tenant summary statistics for the last 30 days
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
		flags.WithPowershellName("Get-TenantUsageSummaryStatistics"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.tenantUsageStatisticsSummary+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListSummaryForTenantCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := usagestatistics.ListSummaryOptions{
			DateFrom: in.TimeValue("dateFrom"),
			DateTo:   in.TimeValue("dateTo"),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.Tenants.UsageStatistics.ListSummary(ctx, opt))
		}, nil
	})
}
