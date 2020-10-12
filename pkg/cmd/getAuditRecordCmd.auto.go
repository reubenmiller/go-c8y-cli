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

type getAuditRecordCmd struct {
	*baseCmd
}

func newGetAuditRecordCmd() *getAuditRecordCmd {
	ccmd := &getAuditRecordCmd{}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get an audit record",
		Long:  `Get an audit record by id`,
		Example: `
$ c8y auditRecords get --id 12345
Get an audit record by id
        `,
		PreRunE: nil,
		RunE:    ccmd.getAuditRecord,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Audit id (required)")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getAuditRecordCmd) getAuditRecord(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
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
	if v, err := cmd.Flags().GetString("id"); err == nil {
		if v != "" {
			pathParameters["id"] = v
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "id", err))
	}

	path := replacePathParameters("/audit/auditRecords/{id}", pathParameters)

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
