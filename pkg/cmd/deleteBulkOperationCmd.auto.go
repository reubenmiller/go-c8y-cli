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

type deleteBulkOperationCmd struct {
	*baseCmd
}

func newDeleteBulkOperationCmd() *deleteBulkOperationCmd {
	ccmd := &deleteBulkOperationCmd{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Remove bulk operation/s. Only bulk operations that are in ACTIVE or IN_PROGRESS can be deleted",
		Long:  ``,
		Example: `
$ c8y bulkOperations delete --id 12345
Remove bulk operation by id
        `,
		PreRunE: validateDeleteMode,
		RunE:    ccmd.deleteBulkOperation,
	}

	cmd.SilenceUsage = true

	cmd.Flags().Int("id", 0, "Bulk Operation id (required)")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *deleteBulkOperationCmd) deleteBulkOperation(cmd *cobra.Command, args []string) error {

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

	// path parameters
	pathParameters := make(map[string]string)
	if v, err := cmd.Flags().GetInt("id"); err == nil {
		pathParameters["id"] = fmt.Sprintf("%d", v)
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "id", err))
	}

	path := replacePathParameters("devicecontrol/bulkoperations/{id}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "DELETE",
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
