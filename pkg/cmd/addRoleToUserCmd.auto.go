// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type AddRoleToUserCmd struct {
	*baseCmd
}

func NewAddRoleToUserCmd() *AddRoleToUserCmd {
	ccmd := &AddRoleToUserCmd{}
	cmd := &cobra.Command{
		Use:   "addRoleTouser",
		Short: "Add role to a user",
		Long:  ``,
		Example: `
$ c8y userRoles addRoleTouser --user "myuser" --role "ROLE_ALARM_READ"
Add a role (ROLE_ALARM_READ) to a user
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().StringSlice("user", []string{""}, "User prefix or full username (required)")
	cmd.Flags().StringSlice("role", []string{""}, "User role id (accepts pipeline)")
	addProcessingModeFlag(cmd)

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("role", "role.self", false, "self", "id"),
	)

	// Required flags
	cmd.MarkFlagRequired("user")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *AddRoleToUserCmd) RunE(cmd *cobra.Command, args []string) error {
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
		return cmderrors.NewUserError(err)
	}

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
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
		return cmderrors.NewUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		WithDataValue(),
		WithRoleSelfByNameFirstMatch(args, "role", "role.self"),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/user/{tenant}/users/{user}/roles")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		flags.WithStringDefaultValue(client.TenantName, "tenant", "tenant"),
		WithUserByNameFirstMatch(args, "user", "user"),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "POST",
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
