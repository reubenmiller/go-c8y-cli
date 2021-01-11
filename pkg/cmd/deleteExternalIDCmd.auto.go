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

type deleteExternalIDCmd struct {
	*baseCmd
}

func newDeleteExternalIDCmd() *deleteExternalIDCmd {
	ccmd := &deleteExternalIDCmd{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete external id",
		Long:  `Delete an existing external id. This does not delete the device managed object`,
		Example: `
$ c8y identity delete --type test --name myserialnumber
Delete external identity
        `,
		PreRunE: validateDeleteMode,
		RunE:    ccmd.deleteExternalID,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("type", "", "External identity type (required)")
	cmd.Flags().String("name", "", "External identity id/name (required)")
	addProcessingModeFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("name")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *deleteExternalIDCmd) deleteExternalID(cmd *cobra.Command, args []string) error {

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
	if v, err := cmd.Flags().GetString("type"); err == nil {
		if v != "" {
			pathParameters["type"] = v
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "type", err))
	}
	if v, err := cmd.Flags().GetString("name"); err == nil {
		if v != "" {
			pathParameters["name"] = v
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "name", err))
	}

	path := replacePathParameters("/identity/externalIds/{type}/{name}", pathParameters)

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
