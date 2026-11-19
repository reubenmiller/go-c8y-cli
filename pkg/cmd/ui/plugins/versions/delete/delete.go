// v2-based UI plugin version delete: the version flag drives iteration (pipe or
// --version) while --plugin (id or name/contextPath, resolved via
// UIPlugins.ResolveID) scopes the request. The selected version is removed by
// --tag (DeleteByTag) or --version (DeleteByVersion). Success yields no output.
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

// NewDeleteCmd creates a command to Delete a specific version of a plugin
func NewDeleteCmd(f *cmdutil.Factory) *DeleteCmd {
	ccmd := &DeleteCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a specific version of a plugin",
		Long:  `Delete a specific version of a plugin in your tenant, by a given tag or version`,
		Example: heredoc.Doc(`
$ c8y ui plugins versions delete --plugin 1234 --tag tag1
Delete plugin version by tag

$ c8y ui plugins versions delete --plugin 1234 --version 1.0
Delete plugin version by version name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("plugin", "", "Plugin")
	cmd.Flags().String("version", "", "Version, e.g. 1.0.0 (accepts pipeline)")
	cmd.Flags().String("tag", "", "The tag of the plugin version")

	completion.WithOptions(
		cmd,
		completion.WithUIPlugin("plugin", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUIPluginVersion("version", "plugin", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("version", "version", false, "id"),
		flags.WithPipelineAliases("plugin", "id"),
		flags.WithPipelineAliases("version", "id"),
		flags.WithPowershellName("Remove-UIPluginVersion"),
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

	if err := r.InputFlag("version"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		pluginID, err := client.UIPlugins.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("plugin")), nil)
		if err != nil {
			return nil, err
		}
		version := in.String("version")
		tag := in.String("tag")
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				if tag != "" {
					return client.ApplicationVersions.DeleteByTag(ctx, pluginID, tag)
				}
				return client.ApplicationVersions.DeleteByVersion(ctx, pluginID, version)
			})
		}, nil
	})
}
