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

type DeleteRoleFromGroupCmd struct {
	*baseCmd
}

func NewDeleteRoleFromGroupCmd() *DeleteRoleFromGroupCmd {
	ccmd := &DeleteRoleFromGroupCmd{}
	cmd := &cobra.Command{
		Use:   "deleteRoleFromGroup",
		Short: "Unassign/Remove role from a group",
		Long:  ``,
		Example: `
$ c8y userRoles deleteRoleFromGroup --group "myuser" --role "ROLE_MEASUREMENT_READ"
Remove a role from the given user
        `,
		PreRunE: validateDeleteMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Group id (required)")
	cmd.Flags().StringSlice("role", []string{""}, "Role name, e.g. ROLE_TENANT_MANAGEMENT_ADMIN (required)")
	cmd.Flags().String("tenant", "", "Tenant")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport(""),
	)

	// Required flags
	cmd.MarkFlagRequired("group")
	cmd.MarkFlagRequired("role")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *DeleteRoleFromGroupCmd) RunE(cmd *cobra.Command, args []string) error {
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
	)
	if err != nil {
		return newUserError(err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	err = flags.WithPathParameters(
		cmd,
		pathParameters,
		flags.WithStringDefaultValue(client.TenantName, "tenant", "tenant"),
	)
	if cmd.Flags().Changed("group") {
		groupInputValues, groupValue, err := getFormattedGroupSlice(cmd, args, "group")

		if err != nil {
			return newUserError("no matching user groups found", groupInputValues, err)
		}

		if len(groupValue) == 0 {
			return newUserError("no matching user groups found", groupInputValues)
		}

		for _, item := range groupValue {
			if item != "" {
				pathParameters["group"] = newIDValue(item).GetID()
			}
		}
	}
	if cmd.Flags().Changed("role") {
		roleInputValues, roleValue, err := getFormattedRoleSlice(cmd, args, "role")

		if err != nil {
			return newUserError("no matching roles found", roleInputValues, err)
		}

		if len(roleValue) == 0 {
			return newUserError("no matching roles found", roleInputValues)
		}

		for _, item := range roleValue {
			if item != "" {
				pathParameters["role"] = newIDValue(item).GetID()
			}
		}
	}

	path := replacePathParameters("/user/{tenant}/groups/{group}/roles/{role}", pathParameters)

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

	return processRequestAndResponseWithWorkers(cmd, &req, PipeOption{"", false})
}
