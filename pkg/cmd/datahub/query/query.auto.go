// Code generated from specification version 1.0.0: DO NOT EDIT
package query

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

// QueryCmd command
type QueryCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewQueryCmd creates a command to Execute a SQL query and retrieve the results
func NewQueryCmd(f *cmdutil.Factory) *QueryCmd {
	ccmd := &QueryCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Execute a SQL query and retrieve the results",
		Long:  `Execute a SQL query and retrieve the results`,
		Example: heredoc.Doc(`
$ c8y datahub query --sql "SELECT * FROM myTenantIdDataLake.Dremio.myTenantId.alarms"
Get a list of alarms from datahub

$ c8y datahub query --sql "SELECT * FROM myTenantIdDataLake.Dremio.myTenantId.alarms" --limit 2000
Get a list of alarms from datahub with custom limit

$ c8y datahub query --sql "SELECT * FROM myTenantIdDataLake.Dremio.myTenantId.alarms" --format PANDAS --raw
Get a list of alarms from datahub using the PANDAS format (note the raw format is necessary here)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("version", "v1", "The version of the high-performance API")
	cmd.Flags().String("sql", "", "The SQL query to execute (accepts pipeline)")
	cmd.Flags().Int("limit", 1000, "The maximum number of query results")
	cmd.Flags().String("format", "", "The response format, which is either DREMIO or PANDAS. The DREMIO format is the same response format as provided by the sql endpoint of the Standard API. The PANDAS format fits to the data format the Pandas library for Python expects.")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("format", "DREMIO", "PANDAS"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("sql", "sql", false, "id"),

		flags.WithSemanticMethod("GET"),
		flags.WithCollectionProperty(".rows[]"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *QueryCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithStringValue("version", "version"),
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
		flags.WithOverrideValue("sql", "sql"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("sql", "sql"),
		flags.WithIntValue("limit", "limit"),
		flags.WithStringValue("format", "format"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("sql", "limit"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("service/datahub/sql")
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
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
