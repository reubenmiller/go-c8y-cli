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

type updateOperationCmd struct {
	*baseCmd
}

func newUpdateOperationCmd() *updateOperationCmd {
	ccmd := &updateOperationCmd{}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update operation",
		Long: `Update an operation. This is commonly used to change an operation's status. For example the operation can be set to FAILED along with a failure reason.
`,
		Example: `
$ c8y operations update --id 12345 --status EXECUTING
Update an operation
		`,
		RunE: ccmd.updateOperation,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Operation id (required)")
	cmd.Flags().String("status", "", "Operation status, can be one of SUCCESSFUL, FAILED, EXECUTING or PENDING. (required)")
	cmd.Flags().String("failureReason", "", "Reason for the failure. Use when setting status to FAILED")
	addDataFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("status")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *updateOperationCmd) updateOperation(cmd *cobra.Command, args []string) error {

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
	if v, err := cmd.Flags().GetString("status"); err == nil {
		if v != "" {
			body.Set("status", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "status", err))
	}
	if v, err := cmd.Flags().GetString("failureReason"); err == nil {
		if v != "" {
			body.Set("failureReason", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "failureReason", err))
	}

	// path parameters
	pathParameters := make(map[string]string)
	if v, err := cmd.Flags().GetString("id"); err == nil {
		if v != "" {
			pathParameters["id"] = v
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "id", err))
	}

	path := replacePathParameters("devicecontrol/operations/{id}", pathParameters)

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
