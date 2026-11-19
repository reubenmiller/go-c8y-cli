// v2-based application binary delete: the binaryId flag drives iteration (pipe or
// --binaryId); each binary of the given --application (id or name, resolved
// internally by the SDK) is removed via Applications.DeleteBinary. Success yields
// no output.
package deleteapplicationbinary

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
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
		Long:  `Delete a binary of an application`,
		Example: heredoc.Doc(`
$ c8y applications deleteApplicationBinary --application 12345 --binaryId 6789
Delete an application binary
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("application", "", "Application id (required)")
	cmd.Flags().StringSlice("binaryId", []string{""}, "Application binary id (required) (accepts pipeline)")
	_ = cmd.MarkFlagRequired("application")

	completion.WithOptions(
		cmd,
		completion.WithHostedApplication("application", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("binaryId", "binaryId", true, "id"),
		flags.WithPipelineAliases("application", "id"),
		flags.WithPowershellName("Remove-ApplicationBinary"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteApplicationBinaryCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("binaryId"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		appRef := c8ystream.NameOrID(in.String("application"))
		binaryID := in.String("binaryId")
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Applications.DeleteBinary(ctx, appRef, binaryID)
			})
		}, nil
	})
}
