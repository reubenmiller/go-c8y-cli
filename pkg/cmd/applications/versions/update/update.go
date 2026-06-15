// v2-based application version update: replaces the tags of a given application
// version via ApplicationVersions.Update. The application flag drives iteration
// (pipe or --application); the application reference (id or name) is resolved via
// Applications.ResolveID, the version is selected by --version.
package update

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// UpdateCmd command
type UpdateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUpdateCmd creates a command to Replace an application version's tags
func NewUpdateCmd(f *cmdutil.Factory) *UpdateCmd {
	ccmd := &UpdateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Replace an application version's tags",
		Long:  `Replaces the tags of a given application version in your tenant`,
		Example: heredoc.Doc(`
$ c8y applications versions update --application 1234 --version 1.0 --tag tag1,latest
Replace application version's tags
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("application", "", "Application (accepts pipeline)")
	cmd.Flags().String("version", "", "Application version")
	cmd.Flags().StringSlice("tag", []string{""}, "Tag assigned to the version. Version tags must be unique across all versions and version fields of application versions")

	completion.WithOptions(
		cmd,
		completion.WithApplicationWithVersions("application", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("application", "application", false, "id", "name"),
		flags.WithCollectionProperty("-"),
		flags.WithPowershellName("Update-ApplicationVersionTag"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.applicationVersion+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateCmd) RunE(cmd *cobra.Command, args []string) error {
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
		tags := in.StringSlice("tag")
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.ApplicationVersion] {
				return client.ApplicationVersions.Update(ctx, appID, version, tags)
			})
		}, nil
	})
}
