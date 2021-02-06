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

type DeleteAssetFromGroupCmd struct {
	*baseCmd
}

func NewDeleteAssetFromGroupCmd() *DeleteAssetFromGroupCmd {
	ccmd := &DeleteAssetFromGroupCmd{}
	cmd := &cobra.Command{
		Use:   "unassignAssetFromGroup",
		Short: "Delete child asset reference",
		Long:  `Unassign an asset (device or group) from a group`,
		Example: `
$ c8y inventoryReferences unassignAssetFromGroup --device 12345 --childDevice 22553
Unassign a child device from its parent device
        `,
		PreRunE: validateDeleteMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Asset id (required) (accepts pipeline)")
	cmd.Flags().StringSlice("childDevice", []string{""}, "Child device")
	cmd.Flags().StringSlice("childGroup", []string{""}, "Child device group")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("group"),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *DeleteAssetFromGroupCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	// query parameters
	query := url.Values{}
	err = flags.WithQueryParameters(
		cmd,
		query,
	)
	if err != nil {
		return newUserError(err)
	}

	queryValue, err := url.QueryUnescape(query.Encode())

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
		WithDeviceByNameFirstMatch(args, "childDevice", "reference"),
		WithDeviceGroupByNameFirstMatch(args, "childGroup", "reference"),
	)
	if err != nil {
		return err
	}

	path := replacePathParameters("inventory/managedObjects/{group}/childAssets/{reference}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "DELETE",
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
		Property:          "",
		Required:          true,
		ResolveByNameType: "device",
		IteratorType:      "path",
	}
	return processRequestAndResponseWithWorkers(cmd, &req, pipeOption)
}
