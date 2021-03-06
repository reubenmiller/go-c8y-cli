// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// NewApplicationCmd command
type NewApplicationCmd struct {
	*baseCmd
}

// NewNewApplicationCmd creates a command to Create application
func NewNewApplicationCmd() *NewApplicationCmd {
	ccmd := &NewApplicationCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create application",
		Long:  `Create a new application using explicit settings`,
		Example: `
$ c8y applications create --name myapp --type HOSTED --key "myapp-key" --contextPath "myapp"
Create a new hosted application
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	addDataFlag(cmd)
	cmd.Flags().String("name", "", "Name of application (required)")
	cmd.Flags().String("key", "", "Shared secret of application (required)")
	cmd.Flags().String("type", "", "Type of application. Possible values are EXTERNAL, HOSTED, MICROSERVICE (required)")
	cmd.Flags().String("availability", "", "Access level for other tenants. Possible values are : MARKET, PRIVATE (default)")
	cmd.Flags().String("contextPath", "", "contextPath of the hosted application. Required when application type is HOSTED")
	cmd.Flags().String("resourcesUrl", "", "URL to application base directory hosted on an external server. Required when application type is HOSTED")
	cmd.Flags().String("resourcesUsername", "", "authorization username to access resourcesUrl")
	cmd.Flags().String("resourcesPassword", "", "authorization password to access resourcesUrl")
	cmd.Flags().String("externalUrl", "", "URL to the external application. Required when application type is EXTERNAL")
	addProcessingModeFlag(cmd)

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("type", "EXTERNAL", "HOSTED", "MICROSERVICE"),
		completion.WithValidateSet("availability", "MARKET", "PRIVATE"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("", "", false),
	)

	// Required flags
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("type")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

// RunE executes the command
func (n *NewApplicationCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
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
		return cmderrors.NewUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		WithDataValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("key", "key"),
		flags.WithStringValue("type", "type"),
		flags.WithStringValue("availability", "availability"),
		flags.WithStringValue("contextPath", "contextPath"),
		flags.WithStringValue("resourcesUrl", "resourcesUrl"),
		flags.WithStringValue("resourcesUsername", "resourcesUsername"),
		flags.WithStringValue("resourcesPassword", "resourcesPassword"),
		flags.WithStringValue("externalUrl", "externalUrl"),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/application/applications")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "POST",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: globalFlagIgnoreAccept,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
