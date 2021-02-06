// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type UpdateAlarmCollectionCmd struct {
	*baseCmd
}

func NewUpdateAlarmCollectionCmd() *UpdateAlarmCollectionCmd {
	ccmd := &UpdateAlarmCollectionCmd{}
	cmd := &cobra.Command{
		Use:   "updateCollection",
		Short: "Update a collection of alarms. Currently only the status of alarms can be changed",
		Long:  `Update the status of a collection of alarms by using a filter`,
		Example: `
$ c8y alarms updateCollection --device mydevice --status ACTIVE --newStatus ACKNOWLEDGED
Update the status of all active alarms on a device to ACKNOWLEDGED
        `,
		PreRunE: validateUpdateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "The ManagedObject that the alarm originated from (accepts pipeline)")
	cmd.Flags().String("status", "", "The status of the alarm: ACTIVE, ACKNOWLEDGED or CLEARED. If status was not appeared, new alarm will have status ACTIVE. Must be upper-case.")
	cmd.Flags().String("severity", "", "The severity of the alarm: CRITICAL, MAJOR, MINOR or WARNING. Must be upper-case.")
	cmd.Flags().Bool("resolved", false, "When set to true only resolved alarms will be removed (the one with status CLEARED), false means alarms with status ACTIVE or ACKNOWLEDGED.")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of alarm occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of alarm occurrence.")
	cmd.Flags().String("newStatus", "", "New status to be applied to all of the matching alarms (required)")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("device"),
	)

	// Required flags
	cmd.MarkFlagRequired("newStatus")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *UpdateAlarmCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	err = flags.WithQueryParameters(
		cmd,
		query,
		WithDeviceByNameFirstMatch(args, "device", "source"),
		flags.WithStringValue("status", "status"),
		flags.WithStringValue("severity", "severity"),
		flags.WithBoolValue("resolved", "resolved", ""),
		flags.WithRelativeTimestamp("dateFrom", "dateFrom", ""),
		flags.WithRelativeTimestamp("dateTo", "dateTo", ""),
	)
	if err != nil {
		return newUserError(err)
	}

	queryValue, err = url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
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
	)
	if err != nil {
		return newUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		WithDataValue(),
		flags.WithStringValue("newStatus", "status"),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
	)
	if err != nil {
		return newUserError(err)
	}

	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	err = flags.WithPathParameters(
		cmd,
		pathParameters,
	)

	path := replacePathParameters("alarm/alarms", pathParameters)

	req := c8y.RequestOptions{
		Method:       "PUT",
		Path:         path,
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	pipeOption := PipeOption{
		Name:              "device",
		Property:          "source",
		Required:          false,
		ResolveByNameType: "device",
	}
	return processRequestAndResponseWithWorkers(cmd, &req, pipeOption)
}
