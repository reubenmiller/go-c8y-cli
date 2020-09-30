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

type copyApplicationCmd struct {
	*baseCmd
}

func newCopyApplicationCmd() *copyApplicationCmd {
	ccmd := &copyApplicationCmd{}

	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy application",
		Long: `A POST request to the 'clone' resource creates a new application based on an already existing one.
The properties are copied to the newly created application. For name, key and context path a 'clone' prefix is added in order to be unique.
If the target application is hosted and has an active version, the new application will have the active version with the same content.
The response contains a representation of the newly created application.
Required role ROLE_APPLICATION_MANAGMENT_ADMIN
`,
		Example: `
$ c8y applications copy --id my-example-app
Copy an existing application
		`,
		RunE: ccmd.copyApplication,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Application id (required)")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *copyApplicationCmd) copyApplication(cmd *cobra.Command, args []string) error {

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
		idInputValues, idValue, err := getApplicationSlice(cmd, args, "id")

		if err != nil {
			return newUserError("no matching applications found", idInputValues, err)
		}

		if len(idValue) == 0 {
			return newUserError("no matching applications found", idInputValues)
		}

		for _, item := range idValue {
			if item != "" {
				pathParameters["id"] = newIDValue(item).GetID()
			}
		}
	}

	path := replacePathParameters("/application/applications/{id}/clone", pathParameters)

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
