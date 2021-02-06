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

type NewManagedObjectChildAssetCmd struct {
	*baseCmd
}

func NewNewManagedObjectChildAssetCmd() *NewManagedObjectChildAssetCmd {
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
		WithDataValue(),
		WithDeviceByNameFirstMatch(args, "newChildDevice", "managedObject.id"),
		WithDeviceGroupByNameFirstMatch(args, "newChildGroup", "managedObject.id"),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
	)
	if err != nil {
		return newUserError(err)
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

	pipeOption := PipeOption{
		Name:              "group",
		Property:          "id",
		Required:          true,
		ResolveByNameType: "devicegroup",
	}
	return processRequestAndResponseWithWorkers(cmd, &req, pipeOption)
}
