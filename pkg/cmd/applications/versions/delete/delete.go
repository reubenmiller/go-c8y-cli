// v2-based application version delete: the application flag drives iteration (pipe
// or --application); for each application (id or name, resolved via
// Applications.ResolveID) the selected version is removed by --tag (DeleteByTag)
// or --version (DeleteByVersion). Success yields no output.
package delete

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

// DeleteCmd command
type DeleteCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteCmd creates a command to Delete a specific version of an application
func NewDeleteCmd(f *cmdutil.Factory) *DeleteCmd {
	ccmd := &DeleteCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a specific version of an application",
		Long:  `Delete a specific version of an application in your tenant, by a given tag or version`,
		Example: heredoc.Doc(`
$ c8y applications versions delete --application 1234 --tag tag1
Delete application version by tag

$ c8y applications versions delete --application 1234 --version 1.0
Delete application version by version name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("application", "", "Application (accepts pipeline)")
	cmd.Flags().String("version", "", "The version field of the application version")
	cmd.Flags().String("tag", "", "The tag of the application version")

	completion.WithOptions(
		cmd,
		completion.WithApplicationWithVersions("application", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("application", "application", false, "id", "name"),
		flags.WithPowershellName("Remove-ApplicationVersion"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("application"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		appID, err := client.Applications.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("application")), nil)
		if err != nil {
			return nil, err
		}
		version := in.String("version")
		tag := in.String("tag")
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				if tag != "" {
					return client.ApplicationVersions.DeleteByTag(ctx, appID, tag)
				}
				return client.ApplicationVersions.DeleteByVersion(ctx, appID, version)
			})
		}, nil
	})
}
