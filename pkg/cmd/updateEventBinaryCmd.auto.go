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

type updateEventBinaryCmd struct {
	*baseCmd
}

func newUpdateEventBinaryCmd() *updateEventBinaryCmd {
	ccmd := &updateEventBinaryCmd{}

	cmd := &cobra.Command{
		Use:   "updateBinary",
		Short: "Update event binary",
		Long: `Update an existing event binary
`,
		Example: `
$ c8y events updateBinary --id 12345 --file ./myfile.log
Update a binary related to an event
		`,
		RunE: ccmd.updateEventBinary,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Event id (required)")
	cmd.Flags().String("file", "", "File to be uploaded as a binary (required)")

	// Required flags
	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("file")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *updateEventBinaryCmd) updateEventBinary(cmd *cobra.Command, args []string) error {

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
	if v, err := cmd.Flags().GetString("id"); err == nil {
		if v != "" {
			pathParameters["id"] = v
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "id", err))
	}

	path := replacePathParameters("event/events/{id}/binaries", pathParameters)

	req := c8y.RequestOptions{
		Method:       "PUT",
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
