// Code generated from specification version 1.0.0: DO NOT EDIT
package set

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// SetCmd command
type SetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewSetCmd creates a command to Set log level of microservice
func NewSetCmd(f *cmdutil.Factory) *SetCmd {
	ccmd := &SetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set log level of microservice",
		Long: `Set configured log level for a package (incl. subpackages), or a specific class.
(This only works for Spring Boot microservices based on Cumulocity Java Microservice SDK)
`,
		Example: heredoc.Doc(`
$ c8y microservices loglevels set --name my-microservice --loggerName org.example --logLevel DEBUG
Set log level of microservice for a package

$ c8y microservices loglevels set --name my-microservice --loggerName org.example.microservice.ClassName --logLevel TRACE
Set log level of microservice for a specific class
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Microservice name (required)")
	cmd.Flags().String("loggerName", "", "Name of the logger: Qualified name of package or class (required)")
	cmd.Flags().String("logLevel", "", "Log level: TRACE | DEBUG | INFO | WARN | ERROR | OFF (required)")

	completion.WithOptions(
		cmd,
		completion.WithMicroservice("name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithMicroserviceLoggers("loggerName", "name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("logLevel", "TRACE", "DEBUG", "INFO", "WARN", "ERROR", "OFF"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),

		flags.WithExtendedPipelineSupport("", "", false),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("loggerName")
	_ = cmd.MarkFlagRequired("logLevel")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *SetCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
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
		inputIterators,
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
		flags.WithDataFlagValue(),
		flags.WithStringValue("logLevel", "configuredLogLevel"),
		cmdutil.WithTemplateValue(cfg),
		flags.WithTemplateVariablesValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/service/{name}/loggers/{loggerName}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("loggerName", "loggerName"),
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
