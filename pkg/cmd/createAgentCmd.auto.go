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

type CreateAgentCmd struct {
	*baseCmd
}

func NewCreateAgentCmd() *CreateAgentCmd {
	ccmd := &CreateAgentCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an agent",
		Long: `Create an agent managed object. An agent is a special device managed object with both the
c8y_IsDevice and com_cumulocity_model_Agent fragments.
`,
		Example: `
$ c8y agents create --name myAgent
Create agent

$ c8y agents create --name myAgent --data "custom_value1=1234"
Create agent with custom properties
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Agent name (required)")
	cmd.Flags().String("type", "", "Agent type")
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport(""),
	)

	// Required flags
	cmd.MarkFlagRequired("name")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *CreateAgentCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	err = flags.WithQueryParameters(
		cmd,
		query,
	)
	if err != nil {
		return newUserError(err)
	}
	err = flags.WithQueryOptions(
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

	err = flags.WithHeaders(
		cmd,
		headers,
	)
	if err != nil {
		return newUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		flags.WithDataValue(FlagDataName),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("type", "type"),
	)
	if err != nil {
		return newUserError(err)
	}

	bodyErr := body.MergeJsonnet(`
{  c8y_IsDevice: {},
  com_cumulocity_model_Agent: {},
}
`, false)
	if bodyErr != nil {
		return newSystemError("Template error. ", bodyErr)
	}
	if err := setLazyDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
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

	path := replacePathParameters("inventory/managedObjects", pathParameters)

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

	return processRequestAndResponseWithWorkers(cmd, &req, PipeOption{"", false})
}
