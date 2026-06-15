// v2-based application binary list: the id flag drives iteration (pipe or --id);
// for each application (id or name, resolved internally by the SDK) the binary
// attachments are listed via Applications.ListBinaries. The attachments are a flat
// (non-paginated) collection, so they are flattened to items directly.
package listapplicationbinaries

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// ListApplicationBinariesCmd command
type ListApplicationBinariesCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListApplicationBinariesCmd creates a command to Get application binary collection
func NewListApplicationBinariesCmd(f *cmdutil.Factory) *ListApplicationBinariesCmd {
	ccmd := &ListApplicationBinariesCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "listApplicationBinaries",
		Short: "Get application binary collection",
		Long:  `Retrieve all binaries of an application`,
		Example: heredoc.Doc(`
$ c8y applications listApplicationBinaries --id 12345
List the binaries of an application
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Application id (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithHostedApplication("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPipelineAliases("id", "id"),
		flags.WithCollectionProperty("attachments"),
		flags.WithPowershellName("Get-ApplicationBinaryCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.customAttachmentCollection+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListApplicationBinariesCmd) RunE(cmd *cobra.Command, args []string) error {
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
			return output.FromIterator(client.Applications.ListBinaries(ctx, ref).Items())
		}, nil
	})
}
