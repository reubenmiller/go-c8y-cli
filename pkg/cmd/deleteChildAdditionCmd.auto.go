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

type deleteChildAdditionCmd struct {
	*baseCmd
}

func newDeleteChildAdditionCmd() *deleteChildAdditionCmd {
	ccmd := &deleteChildAdditionCmd{}

	cmd := &cobra.Command{
		Use:   "unassignChildAddition",
		Short: "Delete child addition reference",
		Long:  `Unassign a child addition from an existing managed object`,
		Example: `
$ c8y inventoryReferences unassignChildAddition --id 12345 --childId 22553
Unassign a child addition from its parent managed object
        `,
		PreRunE: validateDeleteMode,
		RunE:    ccmd.deleteChildAddition,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Managed object id (required)")
	cmd.Flags().String("childId", "", "Child managed object id")
	addProcessingModeFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *deleteChildAdditionCmd) deleteChildAddition(cmd *cobra.Command, args []string) error {

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
	if cmd.Flags().Changed("processingMode") {
		if v, err := cmd.Flags().GetString("processingMode"); err == nil && v != "" {
			headers.Add("X-Cumulocity-Processing-Mode", v)
		}
	}

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
	if v, err := cmd.Flags().GetString("childId"); err == nil {
		if v != "" {
			pathParameters["childId"] = v
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "childId", err))
	}

	path := replacePathParameters("inventory/managedObjects/{id}/childAdditions/{childId}", pathParameters)

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
