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

type updateAlarmCollectionCmd struct {
	*baseCmd
}

func newUpdateAlarmCollectionCmd() *updateAlarmCollectionCmd {
	ccmd := &updateAlarmCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "updateCollection",
		Short: "Update a collection of alarms. Currently only the status of alarms can be changed",
		Long:  `Update the status of a collection of alarms by using a filter`,
		Example: `
$ c8y alarms updateCollection --device mydevice --status ACTIVE --newStatus ACKNOWLEDGED
Update the status of all active alarms on a device to ACKNOWLEDGED
		`,
		RunE: ccmd.updateAlarmCollection,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "The ManagedObject that the alarm originated from")
	cmd.Flags().String("status", "", "The status of the alarm: ACTIVE, ACKNOWLEDGED or CLEARED. If status was not appeared, new alarm will have status ACTIVE. Must be upper-case.")
	cmd.Flags().String("severity", "", "The severity of the alarm: CRITICAL, MAJOR, MINOR or WARNING. Must be upper-case.")
	cmd.Flags().Bool("resolved", false, "When set to true only resolved alarms will be removed (the one with status CLEARED), false means alarms with status ACTIVE or ACKNOWLEDGED.")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of alarm occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of alarm occurrence.")
	cmd.Flags().String("newStatus", "", "New status to be applied to all of the matching alarms (required)")

	// Required flags
	cmd.MarkFlagRequired("newStatus")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *updateAlarmCollectionCmd) updateAlarmCollection(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
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
				query.Add("source", newIDValue(item).GetID())
			}
		}
	}
	if v, err := cmd.Flags().GetString("status"); err == nil {
		if v != "" {
			query.Add("status", url.QueryEscape(v))
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "status", err))
	}
	if v, err := cmd.Flags().GetString("severity"); err == nil {
		if v != "" {
			query.Add("severity", url.QueryEscape(v))
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "severity", err))
	}
	if cmd.Flags().Changed("resolved") {
		if v, err := cmd.Flags().GetBool("resolved"); err == nil {
			query.Add("resolved", fmt.Sprintf("%v", v))
		} else {
			return newUserError("Flag does not exist")
		}
	}
	if flagVal, err := cmd.Flags().GetString("dateFrom"); err == nil && flagVal != "" {
		if v, err := tryGetTimestampFlag(cmd, "dateFrom"); err == nil && v != "" {
			query.Add("dateFrom", v)
		} else {
			return newUserError("invalid date format", err)
		}
	}
	if flagVal, err := cmd.Flags().GetString("dateTo"); err == nil && flagVal != "" {
		if v, err := tryGetTimestampFlag(cmd, "dateTo"); err == nil && v != "" {
			query.Add("dateTo", v)
		} else {
			return newUserError("invalid date format", err)
		}
	}
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
	if v, err := cmd.Flags().GetString("newStatus"); err == nil {
		if v != "" {
			body.Set("status", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "newStatus", err))
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("alarm/alarms", pathParameters)

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
