// v2-based measurements getSeries: the device flag drives iteration (pipe or
// --device); for each device the measurement series matching the filter
// (series/aggregation/date) are fetched via Measurements.ListSeries.
package getseries

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/measurements"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetSeriesCmd command
type GetSeriesCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetSeriesCmd creates a command to Get measurement series
func NewGetSeriesCmd(f *cmdutil.Factory) *GetSeriesCmd {
	ccmd := &GetSeriesCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "getSeries",
		Short: "Get measurement series",
		Long:  `Get a collection of measurements based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y measurements getSeries --device 12345 --series app_Weather.temperature --series app_Weather.barometer --dateFrom "-10min" --dateTo "0s"
Get a list of series [app_Weather.temperature] and [app_Weather.barometer] for device 12345
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device ID (accepts pipeline)")
	cmd.Flags().StringSlice("series", []string{""}, "measurement type and series name, e.g. c8y_AccelerationMeasurement.acceleration")
	cmd.Flags().String("aggregationType", "", "Fragment name from measurement.")
	cmd.Flags().String("dateFrom", "-7d", "Start date or date and time of measurement occurrence. Defaults to last 7 days")
	cmd.Flags().String("dateTo", "0s", "End date or date and time of measurement occurrence. Defaults to the current time")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("aggregationType", "DAILY", "HOURLY", "MINUTELY"),
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "source.id", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Get-MeasurementSeries"),
		flags.WithOutputType("application/json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetSeriesCmd) RunE(cmd *cobra.Command, args []string) error {
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
		opt := measurements.ListSeriesOptions{
			DateFrom:        in.TimeValue("dateFrom"),
			DateTo:          in.TimeValue("dateTo"),
			AggregationType: model.MeasurementAggregationType(in.String("aggregationType")),
		}
		for _, s := range in.StringSlice("series") {
			if s != "" {
				opt.Series = append(opt.Series, s)
			}
		}
		if device := in.String("device"); device != "" {
			opt.Source = managedobjects.DeviceRef(c8ystream.NameOrID(device))
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.Measurements.ListSeries(ctx, opt))
		}, nil
	})
}
