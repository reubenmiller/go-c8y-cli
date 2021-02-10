// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"

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
	cmd.Flags().StringSlice("role", []string{""}, "Role name, e.g. ROLE_TENANT_MANAGEMENT_ADMIN (required) (accepts pipeline)")
	cmd.Flags().String("tenant", "", "Tenant")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("role", "role", true),
	)

	// Required flags
	cmd.MarkFlagRequired("group")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *DeleteRoleFromGroupCmd) RunE(cmd *cobra.Command, args []string) error {
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
	)
	if err != nil {
		return newUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/user/{tenant}/groups/{group}/roles/{role}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		WithUserGroupByNameFirstMatch(args, "group", "group"),
		WithRoleByNameFirstMatch(args, "role", "role"),
		flags.WithStringDefaultValue(client.TenantName, "tenant", "tenant"),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "DELETE",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
