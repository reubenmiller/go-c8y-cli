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
	operationCount := &c8ywaiter.OperationCount{
		Client: client,
	}
	if v, err := cmd.Flags().GetInt64("minimum"); err == nil {
		operationCount.Minimum = v
	}

	if v, err := cmd.Flags().GetInt64("maximum"); err == nil {
		operationCount.Maximum = v
	}

	if v, err := cmd.Flags().GetString("dateFrom"); err == nil {
		operationCount.DateFrom = v
	}

	if v, err := cmd.Flags().GetString("dateTo"); err == nil {
		operationCount.DateTo = v
	}

	if v, err := cmd.Flags().GetString("fragmentType"); err == nil {
		operationCount.FragmentType = v
	}

	if v, err := cmd.Flags().GetString("status"); err == nil {
		operationCount.Status = v
	}

	if v, err := cmd.Flags().GetString("bulkOperationId"); err == nil {
		operationCount.BulkOperationId = v
	}

	return operationCount
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
		Short: "Assert operation count",
		Long: heredoc.Doc(`
			Assert that a device has a specific amount of operations and pass the input untouched

			If the assertion is true, then the input value (stdin or an explicit argument value) will be passed untouched to stdout.
			This is useful if you want to filter a list of devices by whether by a specific entity count, and use the results
			in some downstream command (in the pipeline)

			By default, a failed assertion will not set the exit code to a non-zero value. If you want a non-zero exit code
			in such as case then use the --strict option.
		`),
		Example: heredoc.Doc(`
			$ c8y operations assert count --device 1234 --minimum 1
			# => 1234 (if the ID exists)
			# => <no response> (if the ID does not exist)
			# Assert that a device exists, and has at least 1 operation
			
			$ c8y operations assert count --device 1234 --minimum 5 --maximum 10 --dateFrom -1d --strict
			# Assert that the device with id=1111 should have between 5 and 10 operations (inclusive) in the last day
			# Return an error if not

			$ c8y devices list | c8y operations assert count --maximum 0 --dateFrom -7d
			# Find a list of devices which have not created any operations in the last 7 days
		`),
	}

	cmd.Flags().Int64("maximum", -1, "Maximum operation count (inclusive). A value of -1 will disable this check")
	cmd.Flags().Int64("minimum", -1, "Minimum operation count (inclusive). A value of -1 will disable this check")
	cmd.Flags().String("fragmentType", "", "The type of fragment that must be part of the operation. i.e. c8y_Restart")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of operation.")
	cmd.Flags().String("dateTo", "", "End date or date and time of operation.")
	cmd.Flags().String("status", "", "Operation status, can be one of SUCCESSFUL, FAILED, EXECUTING or PENDING.")
	cmd.Flags().String("bulkOperationId", "", "Bulk operation id. Only retrieve operations related to the given bulk operation.")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "PENDING", "EXECUTING", "SUCCESSFUL", "FAILED"),
	)

	assert := &AssertCount{}
	ccmd.SubCommand = subcommand.NewSubCommand(assertFactory.NewAssertDeviceCmdFactory(cmd, f, assert))

	return ccmd
}
