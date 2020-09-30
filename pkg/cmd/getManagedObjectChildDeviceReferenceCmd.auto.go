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

type getManagedObjectChildDeviceReferenceCmd struct {
	*baseCmd
}

func newGetManagedObjectChildDeviceReferenceCmd() *getManagedObjectChildDeviceReferenceCmd {
	ccmd := &getManagedObjectChildDeviceReferenceCmd{}

	cmd := &cobra.Command{
		Use:   "getChildDevice",
		Short: "Get managed object child device reference",
		Long:  ``,
		Example: `
$ c8y inventoryReferences getChildDevice --device 12345 --reference 12345
Get an existing child device reference
		`,
		RunE: ccmd.getManagedObjectChildDeviceReference,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "ManagedObject id (required)")
	cmd.Flags().StringSlice("reference", []string{""}, "Device reference id (required)")

	// Required flags
	cmd.MarkFlagRequired("device")
	cmd.MarkFlagRequired("reference")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getManagedObjectChildDeviceReferenceCmd) getManagedObjectChildDeviceReference(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	if cmd.Flags().Changed("pageSize") {
		if v, err := cmd.Flags().GetInt("pageSize"); err == nil && v > 0 {
			query.Add("pageSize", fmt.Sprintf("%d", v))
		}
	}

	if cmd.Flags().Changed("withTotalPages") {
		if v, err := cmd.Flags().GetBool("withTotalPages"); err == nil && v {
			query.Add("withTotalPages", "true")
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
	if cmd.Flags().Changed("reference") {
		referenceInputValues, referenceValue, err := getFormattedDeviceSlice(cmd, args, "reference")

		if err != nil {
			return newUserError("no matching devices found", referenceInputValues, err)
		}

		if len(referenceValue) == 0 {
			return newUserError("no matching devices found", referenceInputValues)
		}

		for _, item := range referenceValue {
			if item != "" {
				pathParameters["reference"] = newIDValue(item).GetID()
			}
		}
	}

	path := replacePathParameters("inventory/managedObjects/{device}/childDevices/{reference}", pathParameters)

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
