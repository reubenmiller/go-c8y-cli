// Code generated from specification version 1.0.0: DO NOT EDIT
package deleteapplicationbinary

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// DeleteApplicationBinaryCmd command
type DeleteApplicationBinaryCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteApplicationBinaryCmd creates a command to Delete application binary
func NewDeleteApplicationBinaryCmd(f *cmdutil.Factory) *DeleteApplicationBinaryCmd {
	ccmd := &DeleteApplicationBinaryCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "deleteApplicationBinary",
		Short: "Delete application binary",
		Long: `Remove an application binaries related to the given application
The active version can not be deleted and the server will throw an error if you try.
`,
		Example: heredoc.Doc(`
$ c8y applications deleteApplicationBinary --application 12345 --binaryId 9876
Remove an application binary related to a Hosted (web) application
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("application", "", "Application id (required)")
	cmd.Flags().StringSlice("binaryId", []string{""}, "Application binary id (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithApplication("application", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),

		flags.WithExtendedPipelineSupport("binaryId", "binaryId", true, "id"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("application")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteApplicationBinaryCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
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
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/application/applications/{application}/binaries/{binaryId}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithApplicationByNameFirstMatch(client, args, "application", "application"),
		flags.WithStringSliceValues("binaryId", "binaryId", ""),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "DELETE",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		DryRun:       cfg.DryRun(),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
