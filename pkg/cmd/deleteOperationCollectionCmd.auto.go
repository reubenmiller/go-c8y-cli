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

type DeleteOperationCollectionCmd struct {
	*baseCmd
}

func NewDeleteOperationCollectionCmd() *DeleteOperationCollectionCmd {
	ccmd := &DeleteOperationCollectionCmd{}
	cmd := &cobra.Command{
		Use:   "deleteCollection",
		Short: "Delete a collection of operations",
		Long: `Delete a collection of operations using a set of filter criteria. Be careful when deleting operations. Where possible update operations to FAILED (with a failure reason) instead of deleting them as it is easier to track.
`,
		Example: `
$ c8y operations deleteCollection --device mydevice --status PENDING
Remove all pending operations for a given device
        `,
		PreRunE: validateDeleteMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("agent", []string{""}, "Agent ID")
	cmd.Flags().StringSlice("device", []string{""}, "Device ID (accepts pipeline)")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of operation.")
	cmd.Flags().String("dateTo", "", "End date or date and time of operation.")
	cmd.Flags().String("status", "", "Operation status, can be one of SUCCESSFUL, FAILED, EXECUTING or PENDING.")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("device"),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *DeleteOperationCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
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

	err = flags.WithQueryParameters(
		cmd,
		query,
		flags.WithRelativeTimestamp("dateFrom", "dateFrom", ""),
		flags.WithRelativeTimestamp("dateTo", "dateTo", ""),
		flags.WithStringValue("status", "status"),
	)
	if err != nil {
		return newUserError(err)
	}
	err = flags.WithQueryOptions(
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

	err = flags.WithHeaders(
		cmd,
		headers,
	)
	if err != nil {
		return newUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
	)
	if err != nil {
		return newUserError(err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	err = flags.WithPathParameters(
		cmd,
		pathParameters,
	)

	path := replacePathParameters("devicecontrol/operations", pathParameters)

	req := c8y.RequestOptions{
		Method:       "DELETE",
		Path:         path,
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, PipeOption{"device", false})
}
