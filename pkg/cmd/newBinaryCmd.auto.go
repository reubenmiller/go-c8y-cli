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

type newBinaryCmd struct {
	*baseCmd
}

func newNewBinaryCmd() *newBinaryCmd {
	ccmd := &newBinaryCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "New inventory binary",
		Long:  `Upload a new binary to Cumulocity`,
		Example: `
$ c8y binaries create --file ./output.log
Upload a log file

$ c8y binaries create --file "myConfig.json" --data "c8y_Global={},type=c8y_upload"
Upload a config file and make it globally accessible for all users
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.newBinary,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("file", "", "File to be uploaded as a binary (required)")
	addDataFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("file")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newBinaryCmd) newBinary(cmd *cobra.Command, args []string) error {

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
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("/inventory/binaries", pathParameters)

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
