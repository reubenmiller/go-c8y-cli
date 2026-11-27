package assert

import (
	"context"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ywait"
	assertFactory "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/assert/factory"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/measurements"
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
		Short: "Assert measurement count",
		Long: heredoc.Doc(`
			Assert that a device has a specific amount of measurements and pass the input untouched

			If the assertion is true, then the input value (stdin or an explicit argument value) will be passed untouched to stdout.
			This is useful if you want to filter a list of devices by whether by a specific entity count, and use the results
			in some downstream command (in the pipeline)

			By default, a failed assertion will not set the exit code to a non-zero value. If you want a non-zero exit code
			in such as case then use the --strict option.
		`),
		Example: heredoc.Doc(`
			$ c8y measurements assert count --device 1234 --minimum 1
			# => 1234 (if the ID exists)
			# => <no response> (if the ID does not exist)
			# Assert that a device exists, and has at least 1 measurement
			
			$ c8y measurements assert count --device 1234 --minimum 5 --maximum 10 --dateFrom -1d --strict
			# Assert that the device with id=1111 should have between 5 and 10 measurements (inclusive) in the last day
			# Return an error if not

			$ c8y devices list | c8y measurements assert count --maximum 0 --dateFrom -7d
			# Find a list of devices which have not created any measurements in the last 7 days
		`),
	}

	cmd.Flags().Int64("maximum", -1, "Maximum measurement count (inclusive). A value of -1 will disable this check")
	cmd.Flags().Int64("minimum", -1, "Minimum measurement count (inclusive). A value of -1 will disable this check")
	cmd.Flags().String("type", "", "Measurement type.")
	cmd.Flags().String("valueFragmentType", "", "value fragment type")
	cmd.Flags().String("valueFragmentSeries", "", "value fragment series")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of measurement occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of measurement occurrence.")

	completion.WithOptions(
		cmd,
		completion.WithDeviceMeasurementValueFragmentType("valueFragmentType", "device", ccmd.factory.ClientV2),
		completion.WithDeviceMeasurementValueFragmentSeries("valueFragmentSeries", "device", "valueFragmentType", ccmd.factory.ClientV2),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(assertFactory.New(cmd, f, assertFactory.Assert{
		Target:    assertFactory.Device,
		PassState: false,
		Check: func(client *apiv2.Client, in *c8ystream.Resolver, ref string) c8ywait.Check {
			minimum, _ := cmd.Flags().GetInt64("minimum")
			maximum, _ := cmd.Flags().GetInt64("maximum")
			opt := measurements.ListOptions{
				DateFrom:            in.TimeValue("dateFrom"),
				DateTo:              in.TimeValue("dateTo"),
				Type:                in.String("type"),
				ValueFragmentType:   in.String("valueFragmentType"),
				ValueFragmentSeries: in.String("valueFragmentSeries"),
			}
			opt.PageSize = 1
			// v2.59.0 counted the pages of one item (withTotalPages)
			opt.WithTotalPages = true
			return c8ywait.Count(client, ref, cmderrors.MeasurementCount, minimum, maximum, func(ctx context.Context, id string) (int64, error) {
				opt := opt
				opt.Source = managedobjects.DeviceRef(id)
				// the SDK always sends revert=false; v2.59.0 sent no revert
				return c8ywait.Total(client.Measurements.List(apiv2.WithRequestHook(ctx, c8ystream.DropRevertFalse), opt))
			})
		},
	}))

	return ccmd
}
