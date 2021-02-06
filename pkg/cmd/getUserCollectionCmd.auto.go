// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type GetUserCollectionCmd struct {
	*baseCmd
}

func NewGetUserCollectionCmd() *GetUserCollectionCmd {
	ccmd := &GetUserCollectionCmd{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get a collection of users based on filter parameters",
		Long:  `Get a collection of users based on filter parameters`,
		Example: `
$ c8y users list
Get a list of users
        `,
		PreRunE: nil,
		RunE:    ccmd.RunE,
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

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport(""),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *GetUserCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	// query parameters
	query := url.Values{}
	err = flags.WithQueryParameters(
		cmd,
		query,
		flags.WithStringValue("username", "username"),
		flags.WithStringValue("groups", "groups"),
		flags.WithStringValue("owner", "owner"),
		flags.WithBoolValue("onlyDevices", "onlyDevices", ""),
		flags.WithBoolValue("withSubusersCount", "withSubusersCount", ""),
		flags.WithBoolValue("withApps", "withApps", ""),
		flags.WithBoolValue("withGroups", "withGroups", ""),
		flags.WithBoolValue("withRoles", "withRoles", ""),
	)
	if err != nil {
		return newUserError(err)
	}
	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParameters(&query)

	queryValue, err := url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
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

	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	err = flags.WithPathParameters(
		cmd,
		pathParameters,
		flags.WithStringDefaultValue(client.TenantName, "tenant", "tenant"),
	)
	if err != nil {
		return err
	}

	path := replacePathParameters("/user/{tenant}/users", pathParameters)

	req := c8y.RequestOptions{
		Method:       "GET",
		Path:         path,
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	pipeOption := PipeOption{
		Name:              "",
		Property:          "",
		Required:          false,
		ResolveByNameType: "",
		IteratorType:      "",
	}
	return processRequestAndResponseWithWorkers(cmd, &req, pipeOption)
}
