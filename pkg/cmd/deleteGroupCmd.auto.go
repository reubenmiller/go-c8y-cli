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

type DeleteGroupCmd struct {
	*baseCmd
}

func NewDeleteGroupCmd() *DeleteGroupCmd {
	ccmd := &DeleteGroupCmd{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a new group",
		Long:  ``,
		Example: `
$ c8y userGroups delete --id 12345
Delete a user group
        `,
		PreRunE: validateDeleteMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().StringSlice("id", []string{""}, "Group id (required) (accepts pipeline)")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("id"),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *DeleteGroupCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithStringDefaultValue(client.TenantName, "tenant", "tenant"),
	)

	path := replacePathParameters("/user/{tenant}/groups/{id}", pathParameters)

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
		Name:              "id",
		Property:          "",
		Required:          true,
		ResolveByNameType: "usergroup",
	}
	return processRequestAndResponseWithWorkers(cmd, &req, pipeOption)
}
