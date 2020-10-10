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

type newManagedObjectChildDeviceCmd struct {
	*baseCmd
}

func newNewManagedObjectChildDeviceCmd() *newManagedObjectChildDeviceCmd {
	ccmd := &newManagedObjectChildDeviceCmd{}

	cmd := &cobra.Command{
		Use:   "assignChildDevice",
		Short: "Create a child device reference",
		Long:  `Create a child device reference`,
		Example: `
$ c8y inventoryReferences assignChildDevice --device 12345 --newChild 44235
Assign a device as a child device to an existing device
		`,
		RunE: ccmd.newManagedObjectChildDevice,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device. (required)")
	cmd.Flags().StringSlice("newChild", []string{""}, "New child device (required)")

	// Required flags
	cmd.MarkFlagRequired("device")
	cmd.MarkFlagRequired("newChild")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newManagedObjectChildDeviceCmd) newManagedObjectChildDevice(cmd *cobra.Command, args []string) error {

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
	if cmd.Flags().Changed("newChild") {
		newChildInputValues, newChildValue, err := getFormattedDeviceSlice(cmd, args, "newChild")

		if err != nil {
			return newUserError("no matching devices found", newChildInputValues, err)
		}

		if len(newChildValue) == 0 {
			return newUserError("no matching devices found", newChildInputValues)
		}

		for _, item := range newChildValue {
			if item != "" {
				body.Set("managedObject.id", newIDValue(item).GetID())
			}
		}
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
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

	path := replacePathParameters("inventory/managedObjects/{device}/childDevices", pathParameters)

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
