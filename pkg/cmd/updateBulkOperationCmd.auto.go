// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type UpdateBulkOperationCmd struct {
	*baseCmd
}

func NewUpdateBulkOperationCmd() *UpdateBulkOperationCmd {
	ccmd := &UpdateBulkOperationCmd{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update bulk operation",
		Long:  `Update bulk operation. Making update on a started bulk operation cancels it and creates/schedules a new one.`,
		Example: `
$ c8y bulkOperations update --id 12345 --creationRamp 15
Update an bulk operation
        `,
		PreRunE: validateUpdateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().Int("id", 0, "Bulk Operation id (required) (accepts pipeline)")
	cmd.Flags().Float32("creationRampSec", 0, "Delay between every operation creation. (required)")
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("id"),
	)

	// Required flags
	cmd.MarkFlagRequired("creationRampSec")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *UpdateBulkOperationCmd) RunE(cmd *cobra.Command, args []string) error {
	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	err := flags.WithQueryOptions(
		cmd,
		query,
	)
	if err != nil {
		return newUserError(err)
	}

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
	body := mapbuilder.NewInitializedMapBuilder()
	body.SetMap(getDataFlag(cmd))
	if v, err := cmd.Flags().GetFloat32("creationRampSec"); err == nil {
		body.Set("creationRamp", v)
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "creationRampSec", err))
	}
	if err := setLazyDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("devicecontrol/bulkoperations/{id}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "PUT",
		Path:         path,
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, PipeOption{"id", true})
}
