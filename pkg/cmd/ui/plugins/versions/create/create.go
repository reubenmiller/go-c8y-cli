// v2-based UI plugin version create: uploads a ZIP (multipart) as a new version
// of a plugin via ApplicationVersions.CreateFromFile (which accepts a local path
// or an http(s) URL). The plugin flag drives iteration (pipe or --plugin); the
// plugin reference (id or name/contextPath) is resolved via UIPlugins.ResolveID.
package create

import (
	"context"
	"fmt"
	"net/http"

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

// CreateCmd command
type CreateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateCmd creates a command to Create a new version of a plugin
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new version of a plugin",
		Long:  `Uploaded version and tags can only contain upper and lower case letters, integers and ., +, -. Other characters are prohibited.`,
		Example: heredoc.Doc(`
$ c8y ui plugins versions create --plugin 1234 --file "./testdata/myapp.zip" --version "2.0.0"
Create a new version for a plugin
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("plugin", "", "Plugin (accepts pipeline)")
	cmd.Flags().String("file", "", "The ZIP file to be uploaded (a local path or an http(s) URL)")
	cmd.Flags().String("version", "", "Plugin version (required)")
	cmd.Flags().StringSlice("tags", []string{""}, "List of tags associated to the version")
	cmd.Flags().StringSlice("tag", []string{""}, "List of tags associated to the version")

	completion.WithOptions(
		cmd,
		completion.WithUIPlugin("plugin", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("plugin", "plugin", false, "id", "name"),
		flags.WithPipelineAliases("plugin", "id"),
		flags.WithCollectionProperty("-"),
		flags.WithPowershellName("New-UIPluginVersion"),
		flags.WithOutputType("application/json", ""),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("version")

	flags.MarkDeprecated(cmd, "tags", "please use 'tag' instead")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateCmd) RunE(cmd *cobra.Command, args []string) error {
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
		file := in.String("file")
		if file == "" {
			return nil, fmt.Errorf("--file is required")
		}
		version := in.String("version")
		// Merge the deprecated --tags alias with --tag (both map to the version
		// tags); drop the empty entries the cobra StringSlice default carries.
		tags := nonEmpty(append(in.StringSlice("tags"), in.StringSlice("tag")...))
		return func(ctx context.Context) output.Seq {
			// SubmitUpload (not Submit): the multipart body can't be prepared for
			// the confirmation prompt without blocking. CreateFromFile opens the
			// local path / downloads the URL when the call runs.
			return c8ystream.SubmitUpload(ctx, http.MethodPost, "", func(ctx context.Context) op.Result[jsonmodels.ApplicationVersion] {
				return client.ApplicationVersions.CreateFromFile(ctx, pluginID, file, version, tags)
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
