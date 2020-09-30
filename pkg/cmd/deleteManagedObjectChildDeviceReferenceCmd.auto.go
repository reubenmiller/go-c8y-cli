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

type deleteManagedObjectChildDeviceReferenceCmd struct {
	*baseCmd
}

func newDeleteManagedObjectChildDeviceReferenceCmd() *deleteManagedObjectChildDeviceReferenceCmd {
	ccmd := &deleteManagedObjectChildDeviceReferenceCmd{}

	cmd := &cobra.Command{
		Use:   "unassignChildDevice",
		Short: "Delete child device reference",
		Long:  ``,
		Example: `
$ c8y inventoryReferences unassignChildDevice --device 12345 --childDevice 22553
Unassign a child device from its parent device
		`,
		RunE: ccmd.deleteManagedObjectChildDeviceReference,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "ManagedObject id (required)")
	cmd.Flags().StringSlice("childDevice", []string{""}, "Child device reference (required)")

	// Required flags
	cmd.MarkFlagRequired("device")
	cmd.MarkFlagRequired("childDevice")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *deleteManagedObjectChildDeviceReferenceCmd) deleteManagedObjectChildDeviceReference(cmd *cobra.Command, args []string) error {

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
	if cmd.Flags().Changed("childDevice") {
		childDeviceInputValues, childDeviceValue, err := getFormattedDeviceSlice(cmd, args, "childDevice")

		if err != nil {
			return newUserError("no matching devices found", childDeviceInputValues, err)
		}

		if len(childDeviceValue) == 0 {
			return newUserError("no matching devices found", childDeviceInputValues)
		}

		for _, item := range childDeviceValue {
			if item != "" {
				pathParameters["childDevice"] = newIDValue(item).GetID()
			}
		}
	}

	path := replacePathParameters("inventory/managedObjects/{device}/childDevices/{childDevice}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "DELETE",
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
