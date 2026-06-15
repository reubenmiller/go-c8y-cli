// v2-based DataHub job cancel: the id flag drives iteration (pipe or --id); each
// running job is cancelled via DataHub.Jobs.Cancel.
package cancel

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

// CancelCmd command
type CancelCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCancelCmd creates a command to Cancel a query job
func NewCancelCmd(f *cmdutil.Factory) *CancelCmd {
	ccmd := &CancelCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel a query job given the ID of the Dremio job executing the query",
		Long:  `Cancel a query job given the ID of the Dremio job executing the query`,
		Example: heredoc.Doc(`
$ c8y datahub jobs cancel --id 1234abcd
Cancel a running Dremio job
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "The unique identifier of a Dremio job (required) (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("id", "id", true, "id"),
		flags.WithPowershellName("Stop-DataHubJob"),
		flags.WithOutputType("application/json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CancelCmd) RunE(cmd *cobra.Command, args []string) error {
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

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		id := in.String("id")
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.DataHubJob] {
				return client.DataHub.Jobs.Cancel(ctx, id)
			})
		}, nil
	})
}
