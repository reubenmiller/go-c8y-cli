// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type UpdateTenantOptionBulkCmd struct {
	*baseCmd
}

func NewUpdateTenantOptionBulkCmd() *UpdateTenantOptionBulkCmd {
	ccmd := &UpdateTenantOptionBulkCmd{}
	cmd := &cobra.Command{
		Use:   "updateBulk",
		Short: "Update multiple tenant options in provided category",
		Long:  ``,
		Example: `
$ c8y tenantOptions updateBulk --category "c8y_cli_tests" --data "{\"option5\":0,\"option6\":1"}"
Update multiple tenant options
        `,
		PreRunE: validateUpdateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("category", "", "Tenant Option category (required)")
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("", "", false),
	)

	// Required flags
	cmd.MarkFlagRequired("category")
	cmd.MarkFlagRequired("data")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *UpdateTenantOptionBulkCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	// query parameters
	query := url.Values{}
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
	)
	if err != nil {
		return newUserError(err)
	}

	queryValue, err := url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		flags.WithProcessingModeValue(),
	)
	if err != nil {
		return newUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
	)
	if err != nil {
		return newUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		WithDataValue(),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
	)
	if err != nil {
		return newUserError(err)
	}

	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	path := flags.NewStringTemplate("/tenant/options/{category}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		flags.WithStringValue("category", "category"),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "PUT",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
