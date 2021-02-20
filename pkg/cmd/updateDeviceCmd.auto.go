// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"

	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type UpdateDeviceCmd struct {
	*baseCmd
}

func NewUpdateDeviceCmd() *UpdateDeviceCmd {
	ccmd := &UpdateDeviceCmd{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update device",
		Long:  `Update properties of an existing device`,
		Example: `
$ c8y devices update --id 12345
Update device by id
        `,
		PreRunE: validateUpdateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Device ID (required) (accepts pipeline)")
	cmd.Flags().String("newName", "", "Device name")
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *UpdateDeviceCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
	)
	if err != nil {
		return newUserError(err)
	}

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
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
		inputIterators,
	)
	if err != nil {
		return newUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		WithDataValue(),
		flags.WithStringValue("newName", "name"),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
	)
	if err != nil {
		return newUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects/{id}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		WithDeviceByNameFirstMatch(args, "id", "id"),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "PUT",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: globalFlagIgnoreAccept,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
