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

type getUsersInGroupCmd struct {
	*baseCmd
}

func newGetUsersInGroupCmd() *getUsersInGroupCmd {
	ccmd := &getUsersInGroupCmd{}

	cmd := &cobra.Command{
		Use:   "listGroupMembership",
		Short: "Get all users in a group",
		Long:  ``,
		Example: `
$ c8y userReferences listGroupMembership --group 1
List the users within a user group
		`,
		RunE: ccmd.getUsersInGroup,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Group ID (required)")
	cmd.Flags().String("tenant", "", "Tenant")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getUsersInGroupCmd) getUsersInGroup(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	if cmd.Flags().Changed("pageSize") {
		if v, err := cmd.Flags().GetInt("pageSize"); err == nil && v > 0 {
			query.Add("pageSize", fmt.Sprintf("%d", v))
		}
	}

	if cmd.Flags().Changed("withTotalPages") {
		if v, err := cmd.Flags().GetBool("withTotalPages"); err == nil && v {
			query.Add("withTotalPages", "true")
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
	if cmd.Flags().Changed("id") {
		idInputValues, idValue, err := getFormattedGroupSlice(cmd, args, "id")

		if err != nil {
			return newUserError("no matching user groups found", idInputValues, err)
		}

		if len(idValue) == 0 {
			return newUserError("no matching user groups found", idInputValues)
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

	path := replacePathParameters("/user/{tenant}/groups/{id}/users", pathParameters)

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
