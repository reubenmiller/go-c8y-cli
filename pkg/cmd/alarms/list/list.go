// v2-based alarms list: fills the typed alarms.ListOptions from the flags. The
// severity/status filters are typed enum slices; the source device is resolved
// by go-c8y. Time-keyset pagination (newest first).
package list

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/alarms"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get alarm collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get alarm collection",
		Long:  `Get a collection of alarms based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y alarms list --severity MAJOR --status ACTIVE
Get active MAJOR alarms

$ c8y devices list --type myType | c8y alarms list --severity CRITICAL
List critical alarms for each piped device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Source device id (accepts pipeline)")
	cmd.Flags().String("type", "", "Alarm type")
	cmd.Flags().StringSlice("status", []string{""}, "Comma separated alarm statuses, for example ACTIVE,CLEARED")
	cmd.Flags().String("severity", "", "Alarm severity, for example CRITICAL, MAJOR, MINOR or WARNING")
	cmd.Flags().Bool("resolved", false, "When set to true only resolved alarms will be returned (the ones with status CLEARED)")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of alarm occurrence")
	cmd.Flags().String("dateTo", "", "End date or date and time of alarm occurrence")
	cmd.Flags().String("createdFrom", "", "Start date or date and time of the alarm creation")
	cmd.Flags().String("createdTo", "", "End date or date and time of the alarm creation")
	cmd.Flags().String("lastUpdatedFrom", "", "Start date or date and time of the last update made")
	cmd.Flags().String("lastUpdatedTo", "", "End date or date and time of the last update made")
	cmd.Flags().Bool("withSourceAssets", false, "When set to true also alarms for related source assets will be included in the request. When this parameter is provided a source must be specified")
	cmd.Flags().Bool("withSourceDevices", false, "When set to true also alarms for related source devices will be included in the request. When this parameter is provided a source must be specified")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "ACTIVE", "ACKNOWLEDGED", "CLEARED"),
		completion.WithValidateSet("severity", "CRITICAL", "MAJOR", "MINOR", "WARNING"),
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "source", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("alarms"),
		flags.WithPowershellName("Get-AlarmCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.alarmCollection+json", "application/vnd.com.nsn.cumulocity.alarm+json"),
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
	paginationStrategy := pagination.StrategyKind(r.Config.PaginationStrategy())

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := alarms.ListOptions{
			Resolved:          in.Bool("resolved"),
			DateFrom:          in.TimeValue("dateFrom"),
			DateTo:            in.TimeValue("dateTo"),
			CreatedFrom:       in.TimeValue("createdFrom"),
			CreatedTo:         in.TimeValue("createdTo"),
			LastUpdatedFrom:   in.TimeValue("lastUpdatedFrom"),
			LastUpdatedTo:     in.TimeValue("lastUpdatedTo"),
			WithSourceAssets:  in.Bool("withSourceAssets"),
			WithSourceDevices: in.Bool("withSourceDevices"),
		}
		if t := in.String("type"); t != "" {
			opt.Type = []string{t}
		}
		if s := in.String("severity"); s != "" {
			opt.Severity = []model.AlarmSeverity{model.AlarmSeverity(s)}
		}
		for _, status := range in.StringSlice("status") {
			if status != "" {
				opt.Status = append(opt.Status, model.AlarmStatus(status))
			}
		}
		if device := in.String("device"); device != "" {
			opt.Source = managedobjects.DeviceRef(c8ystream.NameOrID(device))
		}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.Alarms.ListAll), nil
	})
}
