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

type AddGroupToGroupCmd struct {
	*baseCmd
}

func NewAddGroupToGroupCmd() *AddGroupToGroupCmd {
	ccmd := &AddGroupToGroupCmd{}
	cmd := &cobra.Command{
		Use:   "assignGroupToGroup",
		Short: "Add a device group to an existing group",
		Long:  `Assigns a group to a group. The group will be a childAsset of the group`,
		Example: `
$ c8y inventoryReferences assignGroupToGroup --group 12345 --newChildGroup 43234
Add a group to a group

$ c8y inventoryReferences assignGroupToGroup --group 12345 --newChildGroup 43234, 99292, 12222
Add multiple groups to a group
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Group (required)")
	cmd.Flags().StringSlice("newChildGroup", []string{""}, "New child group to be added to the group as an child asset (required) (accepts pipeline)")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("newChildGroup"),
	)

	// Required flags
	cmd.MarkFlagRequired("group")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *AddGroupToGroupCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	err = flags.WithQueryParameters(
		cmd,
		query,
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

	err = flags.WithHeaders(
		cmd,
		headers,
		flags.WithProcessingModeValue(),
	)
	if err != nil {
		return newUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
	)
	if err != nil {
		return newUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		flags.WithDataValue(FlagDataName, ""),
	)
	if err != nil {
		return newUserError(err)
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
	if err := setLazyDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	err = flags.WithPathParameters(
		cmd,
		pathParameters,
	)
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
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, PipeOption{"newChildGroup", true})
}
