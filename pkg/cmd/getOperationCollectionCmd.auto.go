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

type getOperationCollectionCmd struct {
	*baseCmd
}

func newGetOperationCollectionCmd() *getOperationCollectionCmd {
	ccmd := &getOperationCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get a collection of operations based on filter parameters",
		Long:  `Get a collection of operations based on filter parameters`,
		Example: `
$ c8y operations list --status PENDING
Get a list of pending operations

$ c8y operations list --agent myagent --status PENDING
Get a list of pending operations for a given agent and all of its child devices

$ c8y operations list --device mydevice --status PENDING
Get a list of pending operations for a device
        `,
		PreRunE: nil,
		RunE:    ccmd.getOperationCollection,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("agent", []string{""}, "Agent ID")
	cmd.Flags().StringSlice("device", []string{""}, "Device ID")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of operation.")
	cmd.Flags().String("dateTo", "", "End date or date and time of operation.")
	cmd.Flags().String("status", "", "Operation status, can be one of SUCCESSFUL, FAILED, EXECUTING or PENDING.")

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getOperationCollectionCmd) getOperationCollection(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	if cmd.Flags().Changed("agent") {
		agentInputValues, agentValue, err := getFormattedDeviceSlice(cmd, args, "agent")

		if err != nil {
			return newUserError("no matching devices found", agentInputValues, err)
		}

		if len(agentValue) == 0 {
			return newUserError("no matching devices found", agentInputValues)
		}

		for _, item := range agentValue {
			if item != "" {
				query.Add("agentId", newIDValue(item).GetID())
			}
		}
	}
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
				query.Add("deviceId", newIDValue(item).GetID())
			}
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
	if v, err := cmd.Flags().GetString("status"); err == nil {
		if v != "" {
			query.Add("status", url.QueryEscape(v))
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "status", err))
	}
	commonOptions.AddQueryParameters(&query)
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

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("devicecontrol/operations", pathParameters)

	req := c8y.RequestOptions{
		Method:       "GET",
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
