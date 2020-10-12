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

type getRoleReferenceCollectionFromGroupCmd struct {
	*baseCmd
}

func newGetRoleReferenceCollectionFromGroupCmd() *getRoleReferenceCollectionFromGroupCmd {
	ccmd := &getRoleReferenceCollectionFromGroupCmd{}

	cmd := &cobra.Command{
		Use:   "getRoleReferenceCollectionFromGroup",
		Short: "Get collection of user role references from a group",
		Long:  ``,
		Example: `
$ c8y userRoles getRoleReferenceCollectionFromGroup --group "12345"
Get a list of role references for a user group
        `,
		PreRunE: nil,
		RunE:    ccmd.getRoleReferenceCollectionFromGroup,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().StringSlice("group", []string{""}, "Group id (required)")

	// Required flags
	cmd.MarkFlagRequired("group")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getRoleReferenceCollectionFromGroupCmd) getRoleReferenceCollectionFromGroup(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	commonOptions.AddQueryParameters(&query)
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
	if v := getTenantWithDefaultFlag(cmd, "tenant", client.TenantName); v != "" {
		pathParameters["tenant"] = v
	}
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

	path := replacePathParameters("/user/{tenant}/groups/{group}/roles", pathParameters)

	req := c8y.RequestOptions{
		Method:       "GET",
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
