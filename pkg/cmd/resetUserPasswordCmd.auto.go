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

type resetUserPasswordCmd struct {
	*baseCmd
}

func newResetUserPasswordCmd() *resetUserPasswordCmd {
	ccmd := &resetUserPasswordCmd{}

	cmd := &cobra.Command{
		Use:   "resetUserPassword",
		Short: "Reset a user's password",
		Long:  `The password can be reset either by issuing a password reset email (default), or be specifying a new password.`,
		Example: `
$ c8y users resetUserPassword --id "myuser"
Update a user
		`,
		RunE: ccmd.resetUserPassword,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "User id (required)")
	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().String("newPassword", "", "New user password. Min: 6, max: 32 characters. Only Latin1 chars allowed")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *resetUserPasswordCmd) resetUserPassword(cmd *cobra.Command, args []string) error {

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

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewMapBuilder()
	body.SetMap(getDataFlag(cmd))
	if v, err := cmd.Flags().GetString("newPassword"); err == nil {
		if v != "" {
			body.Set("password", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "newPassword", err))
	}
	bodyErr := body.MergeJsonnet(`
addIfEmptyString(base, "password", {sendPasswordResetEmail: true})
`, false)
	if bodyErr != nil {
		return newSystemError("Template error. ", bodyErr)
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	if cmd.Flags().Changed("id") {
		idInputValues, idValue, err := getFormattedUserSlice(cmd, args, "id")

		if err != nil {
			return newUserError("no matching users found", idInputValues, err)
		}

		if len(idValue) == 0 {
			return newUserError("no matching users found", idInputValues)
		}

		for _, item := range idValue {
			if item != "" {
				pathParameters["id"] = newIDValue(item).GetID()
			}
		}
	}
	if v := getTenantWithDefaultFlag(cmd, "tenant", client.TenantName); v != "" {
		pathParameters["tenant"] = v
	}

	path := replacePathParameters("user/{tenant}/users/{id}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "PUT",
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
