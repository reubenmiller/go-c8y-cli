// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// GetMeasurementCollectionCmd command
type GetMeasurementCollectionCmd struct {
	*baseCmd
}

// NewGetMeasurementCollectionCmd creates a command to Get measurement collection
func NewGetMeasurementCollectionCmd() *GetMeasurementCollectionCmd {
	ccmd := &GetMeasurementCollectionCmd{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get measurement collection",
		Long:  `Get a collection of measurements based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y measurements list
Get a list of measurements
        `),
		PreRunE: nil,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device ID (accepts pipeline)")
	cmd.Flags().String("type", "", "Measurement type.")
	cmd.Flags().String("valueFragmentType", "", "value fragment type")
	cmd.Flags().String("valueFragmentSeries", "", "value fragment series")
	cmd.Flags().String("fragmentType", "", "Fragment name from measurement (deprecated).")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of measurement occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of measurement occurrence.")
	cmd.Flags().Bool("revert", false, "Return the newest instead of the oldest measurements. Must be used with dateFrom and dateTo parameters")
	cmd.Flags().Bool("csvFormat", false, "Results will be displayed in csv format. Note: -IncludeAll, is not supported when using using this parameter")
	cmd.Flags().Bool("excelFormat", false, "Results will be displayed in Excel format Note: -IncludeAll, is not supported when using using this parameter")
	cmd.Flags().String("unit", "", "Every measurement fragment which contains 'unit' property will be transformed to use required system of units.")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("unit", "imperial", "metric"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "source", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("measurements"),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetMeasurementCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		WithDeviceByNameFirstMatch(args, "device", "source"),
		flags.WithStringValue("type", "type"),
		flags.WithStringValue("valueFragmentType", "valueFragmentType"),
		flags.WithStringValue("valueFragmentSeries", "valueFragmentSeries"),
		flags.WithStringValue("fragmentType", "fragmentType"),
		flags.WithRelativeTimestamp("dateFrom", "dateFrom", ""),
		flags.WithRelativeTimestamp("dateTo", "dateTo", ""),
		flags.WithBoolValue("revert", "revert", ""),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}
	commonOptions, err := getCommonOptions(cmd)
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
		flags.WithBoolValue("csvFormat", "Accept", "text/csv"),
		flags.WithBoolValue("excelFormat", "Accept", "application/vnd.ms-excel"),
		flags.WithStringValue("unit", "X-Cumulocity-System-Of-Units"),
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
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("measurement/measurements")
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
		IgnoreAccept: cliConfig.IgnoreAcceptHeader(),
		DryRun:       cliConfig.DryRun(),
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
