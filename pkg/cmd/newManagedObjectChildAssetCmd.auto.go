// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type NewManagedObjectChildAssetCmd struct {
	*baseCmd
}

func NewNewManagedObjectChildAssetCmd() *NewManagedObjectChildAssetCmd {
	var _ = fmt.Errorf
	ccmd := &NewManagedObjectChildAssetCmd{}
	cmd := &cobra.Command{
		Use:   "createChildAsset",
		Short: "Add a group or device as an asset to an existing group",
		Long:  `Assigns a group or device to an existing group and marks them as assets`,
		Example: `
$ c8y inventoryReferences createChildAsset --group 12345 --newChildGroup 43234
Create group heirachy (parent group -> child group)
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Group (required) (accepts pipeline)")
	cmd.Flags().StringSlice("newChildDevice", []string{""}, "New child device to be added to the group as an asset")
	cmd.Flags().StringSlice("newChildGroup", []string{""}, "New child device group to be added to the group as an asset")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("group"),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *NewManagedObjectChildAssetCmd) RunE(cmd *cobra.Command, args []string) error {
	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	err := flags.WithQueryOptions(
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

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewInitializedMapBuilder()
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
	if cmd.Flags().Changed("newChildGroup") {
		newChildGroupInputValues, newChildGroupValue, err := getFormattedDeviceGroupSlice(cmd, args, "newChildGroup")

		if err != nil {
			return newUserError("no matching device groups found", newChildGroupInputValues, err)
		}

		if len(newChildGroupValue) == 0 {
			return newUserError("no matching device groups found", newChildGroupInputValues)
		}

		for _, item := range newChildGroupValue {
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

	return processRequestAndResponseWithWorkers(cmd, &req, "group")
}
