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
		`,
		RunE: ccmd.newBinary,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("file", "", "File to be uploaded as a binary (required)")

	// Required flags
	cmd.MarkFlagRequired("file")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newBinaryCmd) newBinary(cmd *cobra.Command, args []string) error {

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
	body.SetMap(getDataFlag(cmd))
	getFileFlag(cmd, "file", formData)

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
