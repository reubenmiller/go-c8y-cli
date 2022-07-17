package count

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ywaiter"
	assertFactory "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/assert/factory"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/desiredstate"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type AssertCount struct{}

func (a *AssertCount) GetStateHandler(cmd *cobra.Command, client *c8y.Client) desiredstate.StateDefiner {
	eventCount := &c8ywaiter.EventCount{
		Client: client,
	}
	if v, err := cmd.Flags().GetInt64("minimum"); err == nil {
		eventCount.Minimum = v
	}

	if v, err := cmd.Flags().GetInt64("maximum"); err == nil {
		eventCount.Maximum = v
	}

	if v, err := cmd.Flags().GetString("dateFrom"); err == nil {
		eventCount.DateFrom = v
	}

	if v, err := cmd.Flags().GetString("dateTo"); err == nil {
		eventCount.DateTo = v
	}

	if v, err := cmd.Flags().GetString("type"); err == nil {
		eventCount.Type = v
	}

	if v, err := cmd.Flags().GetString("fragmentType"); err == nil {
		eventCount.FragmentType = v
	}

	return eventCount
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
		Short: "Assert event count",
		Long: heredoc.Doc(`
			Assert that a device has a specific amount of events and pass the input untouched

			If the assertion is true, then the input value (stdin or an explicit argument value) will be passed untouched to stdout.
			This is useful if you want to filter a list of devices by whether by a specific entity count, and use the results
			in some downstream command (in the pipeline)

			By default, a failed assertion will not set the exit code to a non-zero value. If you want a non-zero exit code
			in such as case then use the --strict option.
		`),
		Example: heredoc.Doc(`
			$ c8y events assert count --device 1234 --minimum 1
			# => 1234 (if the ID exists)
			# => <no response> (if the ID does not exist)
			# Assert that a device exists, and has at least 1 event
			
			$ c8y events assert count --device 1234 --minimum 5 --maximum 10 --dateFrom -1d --strict
			# Assert that the device with id=1111 should have between 5 and 10 events (inclusive) in the last day
			# Return an error if not

			$ c8y devices list | c8y events assert count --maximum 0 --dateFrom -7d
			# Find a list of devices which have not created any events in the last 7 days
		`),
	}

	cmd.Flags().Int64("maximum", -1, "Maximum event count (inclusive). A value of -1 will disable this check")
	cmd.Flags().Int64("minimum", -1, "Minimum event count (inclusive). A value of -1 will disable this check")
	cmd.Flags().String("type", "", "Event type.")
	cmd.Flags().String("fragmentType", "", "Fragment name from event.")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of event occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of event occurrence.")

	assert := &AssertCount{}
	ccmd.SubCommand = subcommand.NewSubCommand(assertFactory.NewAssertDeviceCmdFactory(cmd, f, assert))

	return ccmd
}
