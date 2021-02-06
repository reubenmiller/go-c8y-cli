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

type NewAlarmCmd struct {
	*baseCmd
}

func NewNewAlarmCmd() *NewAlarmCmd {
	ccmd := &NewAlarmCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new alarm",
		Long:  `Create a new alarm on a device or agent.`,
		Example: `
$ c8y alarms create --device mydevice --type c8y_TestAlarm --time "-0s" --text "Test alarm" --severity MAJOR
Create a new alarm for device
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "The ManagedObject that the alarm originated from (required) (accepts pipeline)")
	cmd.Flags().String("type", "", "Identifies the type of this alarm, e.g. 'com_cumulocity_events_TamperEvent'.")
	cmd.Flags().String("time", "0s", "Time of the alarm. Defaults to current timestamp.")
	cmd.Flags().String("text", "", "Text description of the alarm.")
	cmd.Flags().String("severity", "", "The severity of the alarm: CRITICAL, MAJOR, MINOR or WARNING. Must be upper-case.")
	cmd.Flags().String("status", "", "The status of the alarm: ACTIVE, ACKNOWLEDGED or CLEARED. If status was not appeared, new alarm will have status ACTIVE. Must be upper-case.")
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("device"),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *NewAlarmCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	// query parameters
	query := url.Values{}
	err = flags.WithQueryParameters(
		cmd,
		query,
	)
	if err != nil {
		return newUserError(err)
	}

	queryValue, err := url.QueryUnescape(query.Encode())

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
		WithDeviceByNameFirstMatch(args, "device", "source.id"),
		flags.WithStringValue("type", "type"),
		flags.WithRelativeTimestamp("time", "time", ""),
		flags.WithStringValue("text", "text"),
		flags.WithStringValue("severity", "severity"),
		flags.WithStringValue("status", "status"),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
		flags.WithRequiredProperties("type", "text", "time", "severity"),
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
	if err != nil {
		return err
	}

	path := replacePathParameters("alarm/alarms", pathParameters)

	req := c8y.RequestOptions{
		Method:       "POST",
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
		Property:          "source.id",
		Required:          true,
		ResolveByNameType: "device",
		IteratorType:      "body",
	}
	return processRequestAndResponseWithWorkers(cmd, &req, pipeOption)
}
