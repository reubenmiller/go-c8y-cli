// v2-based alarms count: counts the alarms matching the given filter
// (device/date/status/severity) in one call via Alarms.Count. The device flag
// drives iteration (pipe or --device). The result is a plain integer.
package count

import (
	"context"
	"strconv"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/alarms"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// CountCmd command
type CountCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCountCmd creates a command to Retrieve the total number of alarms
func NewCountCmd(f *cmdutil.Factory) *CountCmd {
	ccmd := &CountCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Retrieve the total number of alarms",
		Long:  `Count the total number of active alarms on your tenant`,
		Example: heredoc.Doc(`
$ c8y alarms count --severity MAJOR
Get number of active alarms with the severity set to MAJOR

$ c8y alarms count --dateFrom "-10m" --status ACTIVE
Get number of active alarms which occurred in the last 10 minutes
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Source device id. (accepts pipeline)")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of alarm occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of alarm occurrence.")
	cmd.Flags().String("type", "", "Alarm type.")
	cmd.Flags().StringSlice("status", []string{""}, "Comma separated alarm statuses, for example ACTIVE,CLEARED.")
	cmd.Flags().String("severity", "", "Alarm severity, for example CRITICAL, MAJOR, MINOR or WARNING.")
	cmd.Flags().Bool("resolved", false, "When set to true only resolved alarms will be counted (the one with status CLEARED), false means alarms with status ACTIVE or ACKNOWLEDGED.")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("severity", "CRITICAL", "MAJOR", "MINOR", "WARNING"),
		completion.WithValidateSet("status", "ACTIVE", "ACKNOWLEDGED", "CLEARED"),
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "source.id", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Get-AlarmCount"),
		flags.WithOutputType("text/plain", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CountCmd) RunE(cmd *cobra.Command, args []string) error {
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

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := alarms.CountOptions{
			DateFrom: in.TimeValue("dateFrom"),
			DateTo:   in.TimeValue("dateTo"),
			Resolved: in.Bool("resolved"),
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
		return func(ctx context.Context) output.Seq {
			return func(yield func(jsondoc.JSONDoc, error) bool) {
				res := client.Alarms.Count(ctx, opt)
				if res.Err != nil {
					yield(jsondoc.Empty(), res.Err)
					return
				}
				// Under --dry the request has already been rendered by the dry-run
				// handler; don't also emit a placeholder zero.
				if apiv2.IsDryRun(ctx) {
					return
				}
				yield(jsondoc.New([]byte(strconv.FormatInt(res.Data, 10))), nil)
			}
		}, nil
	})
}
