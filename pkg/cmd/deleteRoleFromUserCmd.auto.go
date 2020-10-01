// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type deleteRoleFromUserCmd struct {
	*baseCmd
}

func newDeleteRoleFromUserCmd() *deleteRoleFromUserCmd {
	ccmd := &deleteRoleFromUserCmd{}

	cmd := &cobra.Command{
		Use:   "deleteRoleFromUser",
		Short: "Unassign/Remove role from a user",
		Long:  ``,
		Example: `
$ c8y userRoles deleteRoleFromUser --user "myuser" --role "ROLE_MEASUREMENT_READ"
Remove a role from the given user
		`,
		RunE: ccmd.deleteRoleFromUser,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("user", []string{""}, "User (required)")
	cmd.Flags().StringSlice("role", []string{""}, "Role name (required)")
	cmd.Flags().String("tenant", "", "Tenant")

	// Required flags
	cmd.MarkFlagRequired("user")
	cmd.MarkFlagRequired("role")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *deleteRoleFromUserCmd) deleteRoleFromUser(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	if cmd.Flags().Changed("pageSize") || globalUseNonDefaultPageSize {
		if v, err := cmd.Flags().GetInt("pageSize"); err == nil && v > 0 {
			query.Add("pageSize", fmt.Sprintf("%d", v))
		}
	}
	queryValue, err = url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewMapBuilder()

	// path parameters
	pathParameters := make(map[string]string)
	if cmd.Flags().Changed("user") {
		userInputValues, userValue, err := getFormattedUserSlice(cmd, args, "user")

		if err != nil {
			return newUserError("no matching users found", userInputValues, err)
		}

		if len(userValue) == 0 {
			return newUserError("no matching users found", userInputValues)
		}

		for _, item := range userValue {
			if item != "" {
				pathParameters["user"] = newIDValue(item).GetID()
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
	if v := getTenantWithDefaultFlag(cmd, "tenant", client.TenantName); v != "" {
		pathParameters["tenant"] = v
	}

	path := replacePathParameters("/user/{tenant}/users/{user}/roles/{role}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "DELETE",
		Path:         path,
		Query:        queryValue,
		Body:         body.GetMap(),
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponse([]c8y.RequestOptions{req}, commonOptions)
}
