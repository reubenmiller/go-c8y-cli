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

type newOperationCmd struct {
	*baseCmd
}

func newNewOperationCmd() *newOperationCmd {
	ccmd := &newOperationCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new operation",
		Long:  `Create a new operation for an agent or device`,
		Example: `
$ c8y operations create --device mydevice --data "{c8y_Restart:{}}"
Create operation for a device
		`,
		RunE: ccmd.newOperation,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Identifies the target device on which this operation should be performed. (required)")
	cmd.Flags().String("description", "", "Text description of the operation.")
	addDataFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("device")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newOperationCmd) newOperation(cmd *cobra.Command, args []string) error {

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
				body.Set("deviceId", newIDValue(item).GetID())
			}
		}
	}
	if v, err := cmd.Flags().GetString("description"); err == nil {
		if v != "" {
			body.Set("description", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "description", err))
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("devicecontrol/operations", pathParameters)

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
