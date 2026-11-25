package assert

import (
	"context"
	"strings"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ywait"
	assertFactory "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/assert/factory"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/alarms"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/spf13/cobra"
)

type CmdExists struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdCount(f *cmdutil.Factory) *CmdExists {
	ccmd := &CmdExists{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "count",
		Short: "Assert alarm count",
		Long: heredoc.Doc(`
			Assert that a device has a specific amount of alarms and pass the input untouched

			If the assertion is true, then the input value (stdin or an explicit argument value) will be passed untouched to stdout.
			This is useful if you want to filter a list of devices by whether by a specific entity count, and use the results
			in some downstream command (in the pipeline)

			By default, a failed assertion will not set the exit code to a non-zero value. If you want a non-zero exit code
			in such as case then use the --strict option.
		`),
		Example: heredoc.Doc(`
			$ c8y alarms assert count --device 1234 --minimum 1
			# => 1234 (if the ID exists)
			# => <no response> (if the ID does not exist)
			# Assert that a device exists, and has at least 1 alarm
			
			$ c8y alarms assert count --device 1234 --minimum 5 --maximum 10 --dateFrom -1d --strict
			# Assert that the device with id=1111 should have between 5 and 10 alarms (inclusive) in the last day
			# Return an error if not

			$ c8y devices list | c8y alarms assert count --maximum 0 --dateFrom -7d
			# Find a list of devices which have not created any alarms in the last 7 days
		`),
	}

	cmd.Flags().Int64("maximum", -1, "Maximum alarm count (inclusive). A value of -1 will disable this check")
	cmd.Flags().Int64("minimum", -1, "Minimum alarm count (inclusive). A value of -1 will disable this check")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of alarm occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of alarm occurrence.")
	cmd.Flags().String("type", "", "Alarm type.")
	cmd.Flags().String("fragmentType", "", "Alarm fragment type.")
	cmd.Flags().String("status", "", "Comma separated alarm statuses, for example ACTIVE,CLEARED.")
	cmd.Flags().String("severity", "", "Alarm severity, for example CRITICAL, MAJOR, MINOR or WARNING.")
	cmd.Flags().Bool("resolved", false, "When set to true only resolved alarms will be removed (the one with status CLEARED), false means alarms with status ACTIVE or ACKNOWLEDGED.")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "ACTIVE", "ACKNOWLEDGED", "CLEARED"),
		completion.WithValidateSet("severity", "CRITICAL", "MAJOR", "MINOR", "WARNING"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(assertFactory.New(cmd, f, assertFactory.Assert{
		Target:    assertFactory.Device,
		PassState: false,
		Check: func(client *apiv2.Client, in *c8ystream.Resolver, ref string) c8ywait.Check {
			minimum, _ := cmd.Flags().GetInt64("minimum")
			maximum, _ := cmd.Flags().GetInt64("maximum")
			opt := alarms.ListOptions{
				DateFrom: in.TimeValue("dateFrom"),
				DateTo:   in.TimeValue("dateTo"),
				Resolved: in.Bool("resolved"),
			}
			opt.PageSize = 1
			opt.WithTotalElements = true
			if v := in.String("type"); v != "" {
				opt.Type = []string{v}
			}
			for _, v := range strings.Split(in.String("status"), ",") {
				if v = strings.TrimSpace(v); v != "" {
					opt.Status = append(opt.Status, model.AlarmStatus(v))
				}
			}
			if v := in.String("severity"); v != "" {
				opt.Severity = []model.AlarmSeverity{model.AlarmSeverity(v)}
			}
			// GAP: --fragmentType has no field in alarms.ListOptions (the alarm API does not document it).
			return c8ywait.Count(client, ref, cmderrors.AlarmCount, minimum, maximum, func(ctx context.Context, id string) (int64, error) {
				opt := opt
				opt.Source = managedobjects.DeviceRef(id)
				return c8ywait.Total(client.Alarms.List(ctx, opt))
			})
		},
	}))

	return ccmd
}
