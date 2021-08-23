package count

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ywaiter"
	assertFactory "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/assert/factory"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/desiredstate"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type AssertCount struct{}

func (a *AssertCount) GetStateHandler(cmd *cobra.Command, client *c8y.Client) desiredstate.StateDefiner {
	alarmCount := &c8ywaiter.AlarmCount{
		Client: client,
	}
	if v, err := cmd.Flags().GetInt64("minimum"); err == nil {
		alarmCount.Minimum = v
	}

	if v, err := cmd.Flags().GetInt64("maximum"); err == nil {
		alarmCount.Maximum = v
	}

	if v, err := cmd.Flags().GetString("dateFrom"); err == nil {
		alarmCount.DateFrom = v
	}

	if v, err := cmd.Flags().GetString("dateTo"); err == nil {
		alarmCount.DateTo = v
	}

	if v, err := cmd.Flags().GetString("type"); err == nil {
		alarmCount.Type = v
	}

	if v, err := cmd.Flags().GetString("fragmentType"); err == nil {
		alarmCount.FragmentType = v
	}

	if v, err := cmd.Flags().GetString("status"); err == nil {
		alarmCount.Status = v
	}

	if v, err := cmd.Flags().GetString("severity"); err == nil {
		alarmCount.Severity = v
	}

	if v, err := cmd.Flags().GetBool("resolved"); err == nil {
		alarmCount.Resolved = v
	}

	return alarmCount
}

func (a *AssertCount) GetValue(v interface{}, input interface{}) []byte {
	if raw, ok := input.([]byte); ok {
		return raw
	}
	return nil
}

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

	assert := &AssertCount{}
	ccmd.SubCommand = subcommand.NewSubCommand(assertFactory.NewAssertDeviceCmdFactory(cmd, f, assert))

	return ccmd
}
