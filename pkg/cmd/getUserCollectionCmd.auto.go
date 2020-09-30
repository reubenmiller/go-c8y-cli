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

type getUserCollectionCmd struct {
	*baseCmd
}

func newGetUserCollectionCmd() *getUserCollectionCmd {
	ccmd := &getUserCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get a collection of users based on filter parameters",
		Long:  `Get a collection of users based on filter parameters`,
		Example: `
$ c8y users list
Get a list of users
		`,
		RunE: ccmd.getUserCollection,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().String("username", "", "prefix or full username")
	cmd.Flags().String("groups", "", "numeric group identifiers separated by commas; result will contain only users which belong to at least one of specified groups")
	cmd.Flags().String("owner", "", "exact username")
	cmd.Flags().Bool("onlyDevices", false, "If set to 'true', result will contain only users created during bootstrap process (starting with 'device_'). If flag is absent (or false) the result will not contain 'device_' users.")
	cmd.Flags().Bool("withSubusersCount", false, "if set to 'true', then each of returned users will contain additional field 'subusersCount' - number of direct subusers (users with corresponding 'owner').")
	cmd.Flags().Bool("withApps", false, "Include applications related to the user")
	cmd.Flags().Bool("withGroups", false, "Include group information")
	cmd.Flags().Bool("withRoles", false, "Include role information")

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getUserCollectionCmd) getUserCollection(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	if v, err := cmd.Flags().GetString("username"); err == nil {
		if v != "" {
			query.Add("username", url.QueryEscape(v))
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "username", err))
	}
	if v, err := cmd.Flags().GetString("groups"); err == nil {
		if v != "" {
			query.Add("groups", url.QueryEscape(v))
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "groups", err))
	}
	if v, err := cmd.Flags().GetString("owner"); err == nil {
		if v != "" {
			query.Add("owner", url.QueryEscape(v))
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "owner", err))
	}
	if cmd.Flags().Changed("onlyDevices") {
		if v, err := cmd.Flags().GetBool("onlyDevices"); err == nil {
			query.Add("onlyDevices", fmt.Sprintf("%v", v))
		} else {
			return newUserError("Flag does not exist")
		}
	}
	if cmd.Flags().Changed("withSubusersCount") {
		if v, err := cmd.Flags().GetBool("withSubusersCount"); err == nil {
			query.Add("withSubusersCount", fmt.Sprintf("%v", v))
		} else {
			return newUserError("Flag does not exist")
		}
	}
	if cmd.Flags().Changed("withApps") {
		if v, err := cmd.Flags().GetBool("withApps"); err == nil {
			query.Add("withApps", fmt.Sprintf("%v", v))
		} else {
			return newUserError("Flag does not exist")
		}
	}
	if cmd.Flags().Changed("withGroups") {
		if v, err := cmd.Flags().GetBool("withGroups"); err == nil {
			query.Add("withGroups", fmt.Sprintf("%v", v))
		} else {
			return newUserError("Flag does not exist")
		}
	}
	if cmd.Flags().Changed("withRoles") {
		if v, err := cmd.Flags().GetBool("withRoles"); err == nil {
			query.Add("withRoles", fmt.Sprintf("%v", v))
		} else {
			return newUserError("Flag does not exist")
		}
	}
	commonOptions.AddQueryParameters(&query)
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
	if v := getTenantWithDefaultFlag(cmd, "tenant", client.TenantName); v != "" {
		pathParameters["tenant"] = v
	}

	path := replacePathParameters("/user/{tenant}/users", pathParameters)

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
