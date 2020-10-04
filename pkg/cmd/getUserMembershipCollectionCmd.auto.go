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

type getUserMembershipCollectionCmd struct {
	*baseCmd
}

func newGetUserMembershipCollectionCmd() *getUserMembershipCollectionCmd {
	ccmd := &getUserMembershipCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "listUserMembership",
		Short: "Get information about all groups that a user is a member of",
		Long:  ``,
		Example: `
$ c8y users listUserMembership --id "myuser"
Get a list of groups that a user belongs to
		`,
		RunE: ccmd.getUserMembershipCollection,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "User (required)")
	cmd.Flags().String("tenant", "", "Tenant")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getUserMembershipCollectionCmd) getUserMembershipCollection(cmd *cobra.Command, args []string) error {

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

	path := replacePathParameters("/user/{tenant}/users/{id}/groups", pathParameters)

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
