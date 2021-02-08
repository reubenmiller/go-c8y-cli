// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type DeleteAlarmCollectionCmd struct {
	*baseCmd
}

func NewDeleteAlarmCollectionCmd() *DeleteAlarmCollectionCmd {
	ccmd := &DeleteAlarmCollectionCmd{}
	cmd := &cobra.Command{
		Use:   "deleteCollection",
		Short: "Delete a collection of alarms",
		Long:  `Delete a collection of alarms by a given filter`,
		Example: `
$ c8y alarms deleteCollection --device mydevice --severity MAJOR
Remove alarms on the device with the severity set to MAJOR

$ c8y alarms deleteCollection --device mydevice --dateFrom "-10m" --status ACTIVE
Remove alarms on the device which are active and created in the last 10 minutes
        `,
		PreRunE: validateDeleteMode,
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
	cmd.Flags().Bool("withSourceAssets", false, "When set to true also alarms for related source assets will be removed. When this parameter is provided also source must be defined.")
	cmd.Flags().Bool("withSourceDevices", false, "When set to true also alarms for related source devices will be removed. When this parameter is provided also source must be defined.")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "source", false),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *DeleteAlarmCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithRelativeTimestamp("dateFrom", "dateFrom", ""),
		flags.WithRelativeTimestamp("dateTo", "dateTo", ""),
		flags.WithStringValue("type", "type"),
		flags.WithStringValue("fragmentType", "fragmentType"),
		flags.WithStringValue("status", "status"),
		flags.WithStringValue("severity", "severity"),
		flags.WithBoolValue("resolved", "resolved", ""),
		flags.WithBoolValue("withSourceAssets", "withSourceAssets", ""),
		flags.WithBoolValue("withSourceDevices", "withSourceDevices", ""),
	)
	if err != nil {
		return newUserError(err)
	}

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		flags.WithProcessingModeValue(),
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
		Method:       "DELETE",
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
