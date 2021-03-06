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

type newMeasurementCmd struct {
	*baseCmd
}

func newNewMeasurementCmd() *newMeasurementCmd {
	ccmd := &newMeasurementCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new measurement",
		Long:  `Create a new measurement`,
		Example: `
$ c8y measurements create --id 12345 --time "0s" --type "myType" --data "{\"c8y_Winding\":{ \"temperature\":{\"value\": 1.2345,\"unit\":\"°C\"}}}"
Create measurement
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.newMeasurement,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "The ManagedObject which is the source of this measurement. (required)")
	cmd.Flags().String("time", "0s", "Time of the measurement. Defaults to current timestamp.")
	cmd.Flags().String("type", "", "The most specific type of this entire measurement.")
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("device")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newMeasurementCmd) newMeasurement(cmd *cobra.Command, args []string) error {

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
	if cmd.Flags().Changed("processingMode") {
		if v, err := cmd.Flags().GetString("processingMode"); err == nil && v != "" {
			headers.Add("X-Cumulocity-Processing-Mode", v)
		}
	}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewMapBuilder()
	body.SetMap(getDataFlag(cmd))
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
				body.Set("source.id", newIDValue(item).GetID())
			}
		}
	}
	if flagVal, err := cmd.Flags().GetString("time"); err == nil && flagVal != "" {
		if v, err := tryGetTimestampFlag(cmd, "time"); err == nil && v != "" {
			body.Set("time", decodeC8yTimestamp(v))
		} else {
			return newUserError("invalid date format", err)
		}
	}
	if v, err := cmd.Flags().GetString("type"); err == nil {
		if v != "" {
			body.Set("type", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "type", err))
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	body.SetRequiredKeys("type", "time")
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("measurement/measurements", pathParameters)

	req := c8y.RequestOptions{
		Method:       "POST",
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
