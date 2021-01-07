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

type addDeviceToGroupCmd struct {
	*baseCmd
}

func newAddDeviceToGroupCmd() *addDeviceToGroupCmd {
	ccmd := &addDeviceToGroupCmd{}

	cmd := &cobra.Command{
		Use:   "assignDeviceToGroup",
		Short: "Add a device to an existing group",
		Long:  `Assigns a device to a group. The device will be a childAsset of the group`,
		Example: `
$ c8y inventoryReferences assignDeviceToGroup --group 12345 --newChildDevice 43234
Add a device to a group

$ c8y inventoryReferences assignDeviceToGroup --group 12345 --newChildDevice 43234, 99292, 12222
Add multiple devices to a group
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.addDeviceToGroup,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Group (required)")
	cmd.Flags().StringSlice("newChildDevice", []string{""}, "New device to be added to the group as an child asset (required)")
	addProcessingModeFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("group")
	cmd.MarkFlagRequired("newChildDevice")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *addDeviceToGroupCmd) addDeviceToGroup(cmd *cobra.Command, args []string) error {

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
	if cmd.Flags().Changed("newChildDevice") {
		newChildDeviceInputValues, newChildDeviceValue, err := getFormattedDeviceSlice(cmd, args, "newChildDevice")

		if err != nil {
			return newUserError("no matching devices found", newChildDeviceInputValues, err)
		}

		if len(newChildDeviceValue) == 0 {
			return newUserError("no matching devices found", newChildDeviceInputValues)
		}

		for _, item := range newChildDeviceValue {
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
				pathParameters["id"] = newIDValue(item).GetID()
			}
		}
	}

	path := replacePathParameters("inventory/managedObjects/{id}/childAssets", pathParameters)

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
