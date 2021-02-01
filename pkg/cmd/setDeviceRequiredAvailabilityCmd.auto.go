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

type SetDeviceRequiredAvailabilityCmd struct {
	*baseCmd
}

func NewSetDeviceRequiredAvailabilityCmd() *SetDeviceRequiredAvailabilityCmd {
	ccmd := &SetDeviceRequiredAvailabilityCmd{}
	cmd := &cobra.Command{
		Use:   "setRequiredAvailability",
		Short: "Set the required availability of a device",
		Long:  `Devices that have not sent any message in the response interval are considered unavailable. Response interval can have value between -32768 and 32767 and any values out of range will be shrink to range borders. Such devices are marked as unavailable (see below) and an unavailability alarm is raised. Devices with a response interval of zero minutes are considered to be under maintenance. No alarm is raised while a device is under maintenance. Devices that do not contain 'c8y_RequiredAvailability' are not monitored.`,
		Example: `
$ c8y inventory setRequiredAvailability --device 12345 --interval 10
Set the required availability of a device by name to 10 minutes
        `,
		PreRunE: validateUpdateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device ID (required) (accepts pipeline)")
	cmd.Flags().Int("interval", 0, "Interval in minutes (required)")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("device"),
	)

	// Required flags
	cmd.MarkFlagRequired("interval")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *SetDeviceRequiredAvailabilityCmd) RunE(cmd *cobra.Command, args []string) error {
	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	err := flags.WithQueryOptions(
		cmd,
		query,
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
	if cmd.Flags().Changed("processingMode") {
		if v, err := cmd.Flags().GetString("processingMode"); err == nil && v != "" {
			headers.Add("X-Cumulocity-Processing-Mode", v)
		}
	}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	body.SetMap(getDataFlag(cmd))
	if v, err := cmd.Flags().GetInt("interval"); err == nil {
		body.Set("c8y_RequiredAvailability.responseInterval", v)
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "interval", err))
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("inventory/managedObjects/{device}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "PUT",
		Path:         path,
		Query:        queryValue,
		Body:         body.GetMap(),
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, "device")
}
