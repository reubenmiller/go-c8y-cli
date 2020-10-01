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

type getMicroserviceBootstrapUserCmd struct {
	*baseCmd
}

func newGetMicroserviceBootstrapUserCmd() *getMicroserviceBootstrapUserCmd {
	ccmd := &getMicroserviceBootstrapUserCmd{}

	cmd := &cobra.Command{
		Use:   "getBootstrapUser",
		Short: "Get microservice bootstrap user",
		Long: `Get the bootstrap user associated to a microservice. The bootstrap user is required when running a microservice locally (i.e. during development)
`,
		Example: `
$ c8y microservices getBootstrapUser --id 12345
Get application bootstrap user by app id

$ c8y microservices getBootstrapUser --id myapp
Get application bootstrap user by app name
		`,
		RunE: ccmd.getMicroserviceBootstrapUser,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Microservice id (required)")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getMicroserviceBootstrapUserCmd) getMicroserviceBootstrapUser(cmd *cobra.Command, args []string) error {

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
	if cmd.Flags().Lookup("id") != nil {
		idInputValues, idValue, err := getMicroserviceSlice(cmd, args, "id")

		if err != nil {
			return newUserError("no matching microservices found", idInputValues, err)
		}

		if len(idValue) == 0 {
			return newUserError("no matching microservices found", idInputValues)
		}

		for _, item := range idValue {
			if item != "" {
				pathParameters["id"] = newIDValue(item).GetID()
			}
		}
	}

	path := replacePathParameters("/application/applications/{id}/bootstrapUser", pathParameters)

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
