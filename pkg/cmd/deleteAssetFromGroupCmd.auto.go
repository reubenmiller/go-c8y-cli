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

type deleteAssetFromGroupCmd struct {
	*baseCmd
}

func newDeleteAssetFromGroupCmd() *deleteAssetFromGroupCmd {
	ccmd := &deleteAssetFromGroupCmd{}

	cmd := &cobra.Command{
		Use:   "unassignAssetFromGroup",
		Short: "Delete child asset reference",
		Long:  `Unassign an asset (device or group) from a group`,
		Example: `
$ c8y inventoryReferences unassignAssetFromGroup --device 12345 --childDevice 22553
Unassign a child device from its parent device
		`,
		RunE: ccmd.deleteAssetFromGroup,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Asset id (required)")
	cmd.Flags().StringSlice("childDevice", []string{""}, "Child device")
	cmd.Flags().StringSlice("childGroup", []string{""}, "Child device group")

	// Required flags
	cmd.MarkFlagRequired("group")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *deleteAssetFromGroupCmd) deleteAssetFromGroup(cmd *cobra.Command, args []string) error {

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

	// path parameters
	pathParameters := make(map[string]string)
	if cmd.Flags().Changed("group") {
		groupInputValues, groupValue, err := getFormattedDeviceSlice(cmd, args, "group")

		if err != nil {
			return newUserError("no matching devices found", groupInputValues, err)
		}

		if len(groupValue) == 0 {
			return newUserError("no matching devices found", groupInputValues)
		}

		for _, item := range groupValue {
			if item != "" {
				pathParameters["group"] = newIDValue(item).GetID()
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
				pathParameters["reference"] = newIDValue(item).GetID()
			}
		}
	}
	if cmd.Flags().Changed("childGroup") {
		childGroupInputValues, childGroupValue, err := getFormattedDeviceGroupSlice(cmd, args, "childGroup")

		if err != nil {
			return newUserError("no matching device groups found", childGroupInputValues, err)
		}

		if len(childGroupValue) == 0 {
			return newUserError("no matching device groups found", childGroupInputValues)
		}

		for _, item := range childGroupValue {
			if item != "" {
				pathParameters["reference"] = newIDValue(item).GetID()
			}
		}
	}

	path := replacePathParameters("inventory/managedObjects/{group}/childAssets/{reference}", pathParameters)

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
