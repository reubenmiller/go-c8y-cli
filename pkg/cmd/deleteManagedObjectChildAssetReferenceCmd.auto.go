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

type deleteManagedObjectChildAssetReferenceCmd struct {
	*baseCmd
}

func newDeleteManagedObjectChildAssetReferenceCmd() *deleteManagedObjectChildAssetReferenceCmd {
	ccmd := &deleteManagedObjectChildAssetReferenceCmd{}

	cmd := &cobra.Command{
		Use:   "unassignAssetFromGroup",
		Short: "Delete child asset reference",
		Long:  ``,
		Example: `
$ c8y inventoryReferences unassignAssetFromGroup --group 12345 --childDevice 22553
Unassign a child device from its parent device
		`,
		RunE: ccmd.deleteManagedObjectChildAssetReference,
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

func (n *deleteManagedObjectChildAssetReferenceCmd) deleteManagedObjectChildAssetReference(cmd *cobra.Command, args []string) error {

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
	if cmd.Flags().Changed("group") {
		groupInputValues, groupValue, err := getFormattedDeviceGroupSlice(cmd, args, "group")

		if err != nil {
			return newUserError("no matching device groups found", groupInputValues, err)
		}

		if len(groupValue) == 0 {
			return newUserError("no matching device groups found", groupInputValues)
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
