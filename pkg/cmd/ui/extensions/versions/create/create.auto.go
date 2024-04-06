// Code generated from specification version 1.0.0: DO NOT EDIT
package create

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// CreateCmd command
type CreateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateCmd creates a command to Create a new version of an extension
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new version of an extension",
		Long:  `Uploaded version and tags can only contain upper and lower case letters, integers and ., +, -. Other characters are prohibited.`,
		Example: heredoc.Doc(`
$ c8y ui extensions versions create --extension 1234 --file "./testdata/myapp.zip" --version "2.0.0"
Create a new version for an extension
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("extension", "", "Extension (accepts pipeline)")
	cmd.Flags().String("file", "", "The ZIP file to be uploaded")
	cmd.Flags().String("version", "", "The JSON file with version information. (required)")
	cmd.Flags().StringSlice("tags", []string{""}, "The JSON file with version information. todo (required)")
	cmd.Flags().String("data", "", "data")

	completion.WithOptions(
		cmd,
		completion.WithUIExtension("extension", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("extension", "extension", false, "id", "name"),
		flags.WithPipelineAliases("extension", "id"),

		flags.WithCollectionProperty("-"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("version")
	_ = cmd.MarkFlagRequired("tags")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	// Runtime flag options
	flags.WithOptions(
		cmd,
		flags.WithRuntimePipelineProperty(),
	)
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
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
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"),
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
		inputIterators, flags.WithOptionBuilder().
			Append(flags.WithFileReader("file", "applicationBinary")).
			Append(flags.WithFormDataProperty("applicationVersion")).
			Append(flags.WithStringValue("version", "version")).
			Append(flags.WithFormDataProperty("applicationVersion")).
			Append(flags.WithStringSliceValues("tags", "tags", "")).
			Append(flags.WithDataValue("data", "data")).
			Append(cmdutil.WithTemplateValue(n.factory)).
			Append(flags.WithTemplateVariablesValue()).
			Build()...,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder(true)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/application/applications/{application}/versions")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithUIExtensionByNameFirstMatch(n.factory, args, "extension", "extension"),
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
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
