// v2-based application copy: the id flag drives iteration (pipe or --id); each
// reference (an application id or name, resolved internally by the SDK) is cloned
// via Applications.Copy (POST /clone).
package copy

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/applications"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// CopyCmd command
type CopyCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCopyCmd creates a command to Copy application
func NewCopyCmd(f *cmdutil.Factory) *CopyCmd {
	ccmd := &CopyCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy application",
		Long:  `Copy/clone an existing application`,
		Example: heredoc.Doc(`
$ c8y applications copy --id 12345
Copy an application
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Application id (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithApplication("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPipelineAliases("id", "id"),
		flags.WithPowershellName("Copy-Application"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.application+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CopyCmd) RunE(cmd *cobra.Command, args []string) error {
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
		ref := c8ystream.NameOrID(in.String("id"))
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Application] {
				return client.Applications.Copy(ctx, ref, applications.CopyOptions{})
			})
		}, nil
	})
}
