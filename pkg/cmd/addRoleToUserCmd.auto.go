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

type addRoleToUserCmd struct {
	*baseCmd
}

func newAddRoleToUserCmd() *addRoleToUserCmd {
	ccmd := &addRoleToUserCmd{}

	cmd := &cobra.Command{
		Use:   "addRoleTouser",
		Short: "Add role to a user",
		Long:  ``,
		Example: `
$ c8y userRoles addRoleTouser --user "myuser" --role "ROLE_ALARM_READ"
Add a role (ROLE_ALARM_READ) to a user
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.addRoleToUser,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().StringSlice("user", []string{""}, "User prefix or full username (required)")
	cmd.Flags().StringSlice("role", []string{""}, "User role id")
	addProcessingModeFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("user")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *addRoleToUserCmd) addRoleToUser(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	queryValue, err = url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	if cmd.Flags().Changed("processingMode") {
		if v, err := cmd.Flags().GetString("processingMode"); err == nil && v != "" {
			headers.Add("X-Cumulocity-Processing-Mode", v)
		}
	}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewMapBuilder()
	body.SetMap(getDataFlag(cmd))
	if cmd.Flags().Changed("role") {
		roleInputValues, roleValue, err := getFormattedRoleSelfSlice(cmd, args, "role")

		if err != nil {
			return newUserError("no matching roles found", roleInputValues, err)
		}

		if len(roleValue) == 0 {
			return newUserError("no matching roles found", roleInputValues)
		}

		for _, item := range roleValue {
			if item != "" {
				body.Set("role.self", newIDValue(item).GetID())
			}
		}
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	if v := getTenantWithDefaultFlag(cmd, "tenant", client.TenantName); v != "" {
		pathParameters["tenant"] = v
	}
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

	path := replacePathParameters("/user/{tenant}/users/{user}/roles", pathParameters)

	req := c8y.RequestOptions{
		Method:       "POST",
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
