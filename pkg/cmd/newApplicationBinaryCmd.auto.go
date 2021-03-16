// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// NewApplicationBinaryCmd command
type NewApplicationBinaryCmd struct {
	*baseCmd
}

// NewNewApplicationBinaryCmd creates a command to New application binary
func NewNewApplicationBinaryCmd() *NewApplicationBinaryCmd {
	ccmd := &NewApplicationBinaryCmd{}
	cmd := &cobra.Command{
		Use:   "createBinary",
		Short: "New application binary",
		Long: `For the applications of type microservice and web application to be available for Cumulocity platform users, a binary zip file must be uploaded.
For the microservice application, the zip file must consist of    * cumulocity.json - file describing the deployment
    * image.tar - executable docker image

For the web application, the zip file must include index.html in the root directory.
`,
		Example: heredoc.Doc(`
$ c8y applications createBinary --id 12345 --file ./helloworld.zip
Upload application microservice binary
        `),
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Application id (required) (accepts pipeline)")
	cmd.Flags().String("file", "", "File to be uploaded as a binary (required)")
	addProcessingModeFlag(cmd)

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("file")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

// RunE executes the command
func (n *NewApplicationBinaryCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithFormDataFileAndInfo("file", "data")...,
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
		WithTemplateValue(),
		WithTemplateVariablesValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/application/applications/{id}/binaries")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		WithApplicationByNameFirstMatch(args, "id", "id"),
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
		IgnoreAccept: cliConfig.IgnoreAcceptHeader(),
		DryRun:       cliConfig.DryRun(),
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
