// Code generated from specification version 1.0.0: DO NOT EDIT
package getseries

import (
	"fmt"
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
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
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDeviceMeasurementSeries("series", "device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("aggregationType", "DAILY", "HOURLY", "MINUTELY"),
	)

	flags.WithOptions(
		cmd,

		flags.WithExtendedPipelineSupport("device", "source", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id", "time", "creationTime", "creationTime", "lastUpdated", "time", "creationTime", "creationTime", "lastUpdated"), flags.WithPipelineAliases("dateFrom", "deviceId", "source.id", "managedObject.id", "id", "time", "creationTime", "creationTime", "lastUpdated", "time", "creationTime", "creationTime", "lastUpdated"), flags.WithPipelineAliases("dateTo", "deviceId", "source.id", "managedObject.id", "id", "time", "creationTime", "creationTime", "lastUpdated", "time", "creationTime", "creationTime", "lastUpdated"),
		flags.WithCollectionProperty("-"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetSeriesCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	// Runtime flag options
	flags.WithOptions(
		cmd,
		flags.WithRuntimePipelineProperty(),
	)
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
		c8yfetcher.WithDeviceByNameFirstMatch(client, args, "device", "source"),
		flags.WithStringSliceValues("series", "series", ""),
		flags.WithStringValue("aggregationType", "aggregationType"),
		flags.WithEncodedRelativeTimestamp("dateFrom", "dateFrom"),
		flags.WithEncodedRelativeTimestamp("dateTo", "dateTo"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}
	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParameters(query)

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder(false)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("measurement/measurements/series")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "GET",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
