// v2-based DataHub job results: the id flag drives iteration (pipe or --id); the
// result rows of each job are fetched via DataHub.Jobs.GetResults and flattened
// into individual documents.
package listresults

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/datahub/jobs"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// ListResultsCmd command
type ListResultsCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListResultsCmd creates a command to Retrieve the query results of a job
func NewListResultsCmd(f *cmdutil.Factory) *ListResultsCmd {
	ccmd := &ListResultsCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "listResults",
		Short: "Retrieve the query results given the ID of the Dremio job that has executed the query",
		Long:  `Retrieve the query results given the ID of the Dremio job that has executed the query`,
		Example: heredoc.Doc(`
$ c8y datahub jobs listResults --id 1234abcd
Get the result rows of a completed Dremio job
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "The unique identifier of a Dremio job (required) (accepts pipeline)")
	cmd.Flags().Int("offset", 0, "The offset of the paginated results")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true, "id"),
		flags.WithCollectionProperty("rows"),
		flags.WithPowershellName("Get-DataHubJobResult"),
		flags.WithOutputType("application/json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListResultsCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	// offset is a single query parameter shared across piped ids.
	offset, _ := cmd.Flags().GetInt("offset")

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		id := in.String("id")
		opt := jobs.ResultsOptions{Offset: offset}
		return func(ctx context.Context) output.Seq {
			return output.FromIterator(client.DataHub.Jobs.GetResults(ctx, id, opt).Items())
		}, nil
	})
}
