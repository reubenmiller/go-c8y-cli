// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type GetAlarmCollectionCmd struct {
	*baseCmd
}

func NewGetAlarmCollectionCmd() *GetAlarmCollectionCmd {
	ccmd := &GetAlarmCollectionCmd{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get a collection of alarms based on filter parameters",
		Long:  `Get a collection of alarms based on filter parameters`,
		Example: `
$ c8y alarms list --severity MAJOR --pageSize 100
Get alarms with the severity set to MAJOR

$ c8y alarms list --dateFrom "-10m" --status ACTIVE
Get collection of active alarms which occurred in the last 10 minutes
        `,
		PreRunE: nil,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Source device id. (accepts pipeline)")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of alarm occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of alarm occurrence.")
	cmd.Flags().String("type", "", "Alarm type.")
	cmd.Flags().String("fragmentType", "", "Alarm fragment type.")
	cmd.Flags().String("status", "", "Comma separated alarm statuses, for example ACTIVE,CLEARED.")
	cmd.Flags().String("severity", "", "Alarm severity, for example CRITICAL, MAJOR, MINOR or WARNING.")
	cmd.Flags().Bool("resolved", false, "When set to true only resolved alarms will be removed (the one with status CLEARED), false means alarms with status ACTIVE or ACKNOWLEDGED.")
	cmd.Flags().Bool("withAssets", false, "Include assets")
	cmd.Flags().Bool("withDevices", false, "Include devices")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "source", false),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *GetAlarmCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	// query parameters
	query := url.Values{}
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		WithDeviceByNameFirstMatch(args, "device", "source"),
		flags.WithRelativeTimestamp("dateFrom", "dateFrom", ""),
		flags.WithRelativeTimestamp("dateTo", "dateTo", ""),
		flags.WithStringValue("type", "type"),
		flags.WithStringValue("fragmentType", "fragmentType"),
		flags.WithStringValue("status", "status"),
		flags.WithStringValue("severity", "severity"),
		flags.WithBoolValue("resolved", "resolved", ""),
		flags.WithBoolValue("withAssets", "withAssets", ""),
		flags.WithBoolValue("withDevices", "withDevices", ""),
	)
	if err != nil {
		return newUserError(err)
	}
	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParameters(&query)

	queryValue, err := url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
	)
	if err != nil {
		return newUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
	)
	if err != nil {
		return newUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
	)
	if err != nil {
		return newUserError(err)
	}

	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	path := flags.NewStringTemplate("alarm/alarms")
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
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
