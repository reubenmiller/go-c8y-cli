// v2-based DataHub job create: the sql flag drives iteration (pipe or --sql);
// each query is submitted as an asynchronous Dremio job via DataHub.Jobs.Create.
package create

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// CreateCmd command
type CreateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateCmd creates a command to Submit a SQL query as a Dremio job
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Submit a SQL query and retrieve the ID of the Dremio job executing this query",
		Long:  `Submit a SQL query and retrieve the ID of the Dremio job executing this query. The request is asynchronous, i.e., the response does not wait for the query execution to complete.`,
		Example: heredoc.Doc(`
$ c8y datahub jobs create --sql "SELECT * FROM mytable"
Submit a SQL query and get the Dremio job id
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("sql", "", "The SQL query to execute (accepts pipeline)")
	cmd.Flags().StringSlice("context", []string{""}, "The context in which the query is executed")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("sql", "sql", false, "id"),
		flags.WithPowershellName("New-DataHubJob"),
		flags.WithOutputType("application/json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithStringSliceValues("context", "context", ""),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("sql"),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.DataHubJob] {
				return client.DataHub.Jobs.Create(ctx, body)
			})
		}, nil
	})
}
