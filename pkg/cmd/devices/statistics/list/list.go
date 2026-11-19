// v2-based device statistics list: the device flag drives iteration (or a single
// run for the whole tenant when unset). Each run fetches daily or monthly device
// statistics for the tenant via DeviceStatistics.ListAllDaily/ListAllMonthly,
// streaming the paged "statistics" items.
package list

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/devicestatistics"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Retrieve device statistics
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Retrieve device statistics",
		Long: `Retrieve device statistics from a specific tenant (by a given ID). Either daily or monthly.
`,
		Example: heredoc.Doc(`
$ c8y devices statistics list
Get daily (default) device statistics for all devices in the current tenant

$ c8y devices statistics list --date "-7d" --type daily
Get daily device statistics for all devices in the current tenant 7 days ago

$ c8y devices statistics list --date "-30d" --device 12345
Get daily device statistics for all devices in the current tenant 30 days ago

$ c8y devices statistics list --date 2022-01-01 --type monthly
Get monthly device statistics for all devices for a specific month (day is ignored)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("date", "-1d", "Date of the queried day. When type is set to monthly then will be ignored.")
	cmd.Flags().String("type", "daily", "Aggregation type. e.g. daily or monthly")
	cmd.Flags().String("tenant", "", "Tenant id. Defaults to current tenant (based on credentials)")
	cmd.Flags().StringSlice("device", []string{""}, "The ID of the device to search for. (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("type", "daily", "monthly"),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "deviceId", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("date", "time", "creationTime", "lastUpdated"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("statistics"),
		flags.WithPowershellName("Get-DeviceStatisticsCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.tenantusagestatisticscollection+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("device"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	common, err := r.Config.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}
	rawOutput := r.Config.RawOutput()

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		deviceID := ""
		if dev := in.String("device"); dev != "" {
			deviceID, err = client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(dev), nil)
			if err != nil {
				return nil, err
			}
		}

		tenant := in.String("tenant")
		if tenant == "" {
			tenant = n.factory.GetTenant()
		}

		opt := devicestatistics.ListOptions{
			TenantID: tenant,
			Date:     in.TimeValue("date").Format("2006-01-02"),
			DeviceID: deviceID,
			PaginationOptions: pagination.PaginationOptions{
				PageSize:          common.PageSize,
				WithTotalPages:    common.WithTotalPages,
				WithTotalElements: common.WithTotalElements,
				CurrentPage:       int(common.CurrentPage),
				MaxItems:          r.Config.MaxItems(),
			},
		}

		listAll := client.DeviceStatistics.ListAllDaily
		if in.String("type") == "monthly" {
			listAll = client.DeviceStatistics.ListAllMonthly
		}
		return c8ystream.ListCall(rawOutput, opt, listAll), nil
	})
}
