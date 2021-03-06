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

type getBulkOperationCollectionCmd struct {
	*baseCmd
}

func newGetBulkOperationCollectionCmd() *getBulkOperationCollectionCmd {
	ccmd := &getBulkOperationCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get a collection of bulk operations",
		Long:  `Get a collection of bulk operations`,
		Example: `
$ c8y bulkOperations list
Get a list of bulk operations
        `,
		PreRunE: nil,
		RunE:    ccmd.getBulkOperationCollection,
	}

	cmd.SilenceUsage = true

	cmd.Flags().Bool("withDeleted", false, "Include CANCELLED bulk operations")

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getBulkOperationCollectionCmd) getBulkOperationCollection(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	if cmd.Flags().Changed("withDeleted") {
		if v, err := cmd.Flags().GetBool("withDeleted"); err == nil {
			query.Add("withDeleted", fmt.Sprintf("%v", v))
		} else {
			return newUserError("Flag does not exist")
		}
	}
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

	path := replacePathParameters("devicecontrol/bulkoperations", pathParameters)

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
