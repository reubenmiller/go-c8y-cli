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

// UpdateEventBinaryCmd command
type UpdateEventBinaryCmd struct {
	*baseCmd
}

// NewUpdateEventBinaryCmd creates a command to Update event binary
func NewUpdateEventBinaryCmd() *UpdateEventBinaryCmd {
	ccmd := &UpdateEventBinaryCmd{}
	cmd := &cobra.Command{
		Use:   "updateBinary",
		Short: "Update event binary",
		Long: `Update an existing event binary
`,
		Example: heredoc.Doc(`
$ c8y events updateBinary --id 12345 --file ./myfile.log
Update a binary related to an event
        `),
		PreRunE: validateUpdateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Event id (required) (accepts pipeline)")
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
func (n *UpdateEventBinaryCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithFilePath("file", "file", ""),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("event/events/{id}/binaries")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		flags.WithStringValue("id", "id"),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "PUT",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body.GetFileContents(),
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: globalFlagIgnoreAccept,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
