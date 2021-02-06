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

type NewApplicationCmd struct {
	*baseCmd
}

func NewNewApplicationCmd() *NewApplicationCmd {
	ccmd := &NewApplicationCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new application",
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

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport(""),
	)

	// Required flags
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("type")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *NewApplicationCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	// query parameters
	query := url.Values{}
	err = flags.WithQueryParameters(
		cmd,
		query,
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
	)
	if err != nil {
		return newUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
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
		return newUserError(err)
	}

	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	err = flags.WithPathParameters(
		cmd,
		pathParameters,
	)
	if err != nil {
		return err
	}

	path := replacePathParameters("/application/applications", pathParameters)

	req := c8y.RequestOptions{
		Method:       "POST",
		Path:         path,
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	pipeOption := PipeOption{
		Name:              "",
		Property:          "",
		Required:          false,
		ResolveByNameType: "",
		IteratorType:      "",
	}
	return processRequestAndResponseWithWorkers(cmd, &req, pipeOption)
}
