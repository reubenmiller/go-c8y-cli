// v2-based event binary delete: the id flag drives iteration (pipe or --id);
// each event's binary is removed via Events.Binaries.Delete. Success yields no
// output; errors flow through the shared pipeline.
package deletebinary

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteBinaryCmd command
type DeleteBinaryCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteBinaryCmd creates a command to Delete event binary
func NewDeleteBinaryCmd(f *cmdutil.Factory) *DeleteBinaryCmd {
	ccmd := &DeleteBinaryCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "deleteBinary",
		Short: "Delete event binary",
		Long:  `Delete a binary which has been attached to an event`,
		Example: heredoc.Doc(`
$ c8y events deleteBinary --id 12345
Delete a binary attached to an event
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Event id (required) (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("Remove-EventBinary"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteBinaryCmd) RunE(cmd *cobra.Command, args []string) error {
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
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Events.Binaries.Delete(ctx, id)
			})
		}, nil
	})
}
