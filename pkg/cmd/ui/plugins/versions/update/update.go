// v2-based UI plugin version update: replaces the tags of a given plugin version
// via ApplicationVersions.Update. The plugin flag drives iteration (pipe or
// --plugin); the plugin reference (id or name/contextPath) is resolved via
// UIPlugins.ResolveID, the version is selected by --version.
package update

import (
	"context"
	"fmt"

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

// NewUpdateCmd creates a command to Replace tags related to a plugin version
func NewUpdateCmd(f *cmdutil.Factory) *UpdateCmd {
	ccmd := &UpdateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Replace tags related to a plugin version",
		Long:  `Replaces the tags of a given plugin version in your tenant`,
		Example: heredoc.Doc(`
$ c8y ui plugins versions update --plugin 1234 --version 1.0 --tag tag1,latest
Replace tags assigned to a version of a plugin
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("plugin", "", "Plugin (accepts pipeline)")
	cmd.Flags().String("version", "", "Version")
	cmd.Flags().StringSlice("tags", []string{""}, "Tag assigned to the version. Version tags must be unique across all versions and version fields of plugin versions")
	cmd.Flags().StringSlice("tag", []string{""}, "Tag assigned to the version. Version tags must be unique across all versions and version fields of plugin versions")

	completion.WithOptions(
		cmd,
		completion.WithUIPlugin("plugin", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUIPluginVersion("version", "plugin", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("plugin", "plugin", false, "id", "name"),
		flags.WithPipelineAliases("plugin", "id"),
		flags.WithPipelineAliases("version", "id"),
		flags.WithCollectionProperty("-"),
		flags.WithPowershellName("Update-UIPluginVersion"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.applicationVersion+json", ""),
	)

	flags.MarkDeprecated(cmd, "tags", "please use 'tag' instead")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("plugin"); err != nil {
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
		// Merge the deprecated --tags alias with --tag; drop the empty entries the
		// cobra StringSlice default carries. At least one tag is required (the
		// endpoint replaces the full tag set).
		tags := nonEmpty(append(in.StringSlice("tags"), in.StringSlice("tag")...))
		if len(tags) == 0 {
			return nil, fmt.Errorf("at least one --tag is required")
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.ApplicationVersion] {
				return client.ApplicationVersions.Update(ctx, pluginID, version, tags)
			})
		}, nil
	})
}

// nonEmpty returns the input slice without empty-string entries.
func nonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
