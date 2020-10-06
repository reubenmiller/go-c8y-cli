// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type setDeviceRequiredAvailabilityCmd struct {
	*baseCmd
}

func newSetDeviceRequiredAvailabilityCmd() *setDeviceRequiredAvailabilityCmd {
	ccmd := &setDeviceRequiredAvailabilityCmd{}

	cmd := &cobra.Command{
		Use:   "setRequiredAvailability",
		Short: "Set the required availability of a device",
		Long:  `Devices that have not sent any message in the response interval are considered unavailable. Response interval can have value between -32768 and 32767 and any values out of range will be shrink to range borders. Such devices are marked as unavailable (see below) and an unavailability alarm is raised. Devices with a response interval of zero minutes are considered to be under maintenance. No alarm is raised while a device is under maintenance. Devices that do not contain 'c8y_RequiredAvailability' are not monitored.`,
		Example: `
$ c8y inventory setRequiredAvailability --device 12345 --interval 10
Set the required availability of a device by name to 10 minutes
		`,
		RunE: ccmd.setDeviceRequiredAvailability,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device ID (required)")
	cmd.Flags().Int("interval", 0, "Interval in minutes (required)")

	// Required flags
	cmd.MarkFlagRequired("device")
	cmd.MarkFlagRequired("interval")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *setDeviceRequiredAvailabilityCmd) setDeviceRequiredAvailability(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	queryValue, err = url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewMapBuilder()
	body.SetMap(getDataFlag(cmd))
	if v, err := cmd.Flags().GetInt("interval"); err == nil {
		body.Set("c8y_RequiredAvailability.responseInterval", v)
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "interval", err))
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	if cmd.Flags().Changed("device") {
		deviceInputValues, deviceValue, err := getFormattedDeviceSlice(cmd, args, "device")

		if err != nil {
			return newUserError("no matching devices found", deviceInputValues, err)
		}

		if len(deviceValue) == 0 {
			return newUserError("no matching devices found", deviceInputValues)
		}

		for _, item := range deviceValue {
			if item != "" {
				pathParameters["device"] = newIDValue(item).GetID()
			}
		}
	}

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

	return processRequestAndResponse([]c8y.RequestOptions{req}, commonOptions)
}
