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

type DeleteUserFromGroupCmd struct {
	*baseCmd
}

func NewDeleteUserFromGroupCmd() *DeleteUserFromGroupCmd {
	ccmd := &DeleteUserFromGroupCmd{}
	cmd := &cobra.Command{
		Use:   "deleteUserFromGroup",
		Short: "Delete a user from a group",
		Long:  ``,
		Example: `
$ c8y userReferences deleteUserFromGroup --group 1 --user myuser
List the users within a user group
        `,
		PreRunE: validateDeleteMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Group ID (required)")
	cmd.Flags().StringSlice("user", []string{""}, "User id/username (required)")
	cmd.Flags().String("tenant", "", "Tenant")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport(""),
	)

	// Required flags
	cmd.MarkFlagRequired("group")
	cmd.MarkFlagRequired("user")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *DeleteUserFromGroupCmd) RunE(cmd *cobra.Command, args []string) error {
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
		WithUserGroupByNameFirstMatch(args, "group", "group"),
		WithUserByNameFirstMatch(args, "user", "user"),
		flags.WithStringDefaultValue(client.TenantName, "tenant", "tenant"),
	)

	path := replacePathParameters("/user/{tenant}/groups/{group}/users/{user}", pathParameters)

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
		Name:              "",
		Property:          "",
		Required:          false,
		ResolveByNameType: "",
	}
	return processRequestAndResponseWithWorkers(cmd, &req, pipeOption)
}
