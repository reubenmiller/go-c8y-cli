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

type newApplicationBinaryCmd struct {
	*baseCmd
}

func newNewApplicationBinaryCmd() *newApplicationBinaryCmd {
	ccmd := &newApplicationBinaryCmd{}

	cmd := &cobra.Command{
		Use:   "createBinary",
		Short: "New application binary",
		Long: `For the applications of type microservice and web application to be available for Cumulocity platform users, a binary zip file must be uploaded.
For the microservice application, the zip file must consist of    * cumulocity.json - file describing the deployment
    * image.tar - executable docker image

For the web application, the zip file must include index.html in the root directory.
`,
		Example: `
$ c8y applications createBinary --id 12345 --file ./helloworld.zip
Upload application microservice binary
		`,
		RunE: ccmd.newApplicationBinary,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Application id (required)")
	cmd.Flags().String("file", "", "File to be uploaded as a binary (required)")

	// Required flags
	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("file")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newApplicationBinaryCmd) newApplicationBinary(cmd *cobra.Command, args []string) error {

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
	getFileFlag(cmd, "file", formData)

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

	path := replacePathParameters("/application/applications/{id}/binaries", pathParameters)

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
