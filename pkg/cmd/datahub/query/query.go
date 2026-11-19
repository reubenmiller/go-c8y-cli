// v2-based DataHub SQL query: the sql flag drives iteration (pipe or --sql); each
// query is executed via DataHub.Query and the result rows are flattened into
// individual documents.
package query

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
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
$ c8y datahub query --sql "SELECT * FROM mytable" --limit 10
Execute a SQL query and return the first 10 rows
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
	cmd.Flags().String("format", "", "The response format, which is either DREMIO or PANDAS")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("sql", "sql", false, "id"),
		flags.WithCollectionProperty("rows"),
		flags.WithPowershellName("Get-DataHubQueryResult"),
		flags.WithOutputType("application/json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *QueryCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("sql"); err != nil {
		return err
	}

	err = r.Body(
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
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		version := in.String("version")
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return output.FromIterator(client.DataHub.Query(ctx, version, body).Items())
		}, nil
	})
}
