// v2-based UI plugin version list: the plugin flag drives iteration (pipe or
// --plugin); each plugin reference (id or name/contextPath, resolved via
// UIPlugins.ResolveID) drives a typed ApplicationVersions.ListAll call whose
// pages stream through the shared output pipeline. UI plugin versions are
// application versions (same endpoint), so the ApplicationVersions service backs it.
package list

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	appversions "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/applications/versions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get plugin version collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get plugin version collection",
		Long:  `Get a collection of plugin versions by a given filter`,
		Example: heredoc.Doc(`
$ c8y ui plugins versions list --plugin 1234 --pageSize 100
Get plugin versions

$ c8y ui plugins versions list --plugin mychart
Get versions by plugin name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("plugin", "", "Plugin (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithUIPlugin("plugin", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("plugin", "plugin", false, "id", "name"),
		flags.WithPipelineAliases("plugin", "id"),
		flags.WithCollectionProperty("applicationVersions"),
		flags.WithPowershellName("Get-UIPluginVersionCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.applicationVersionCollection+json", "application/vnd.com.nsn.cumulocity.applicationVersion+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
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

	common, err := r.Config.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	rawOutput := r.Config.RawOutput()
	paginationStrategy := pagination.StrategyKind(r.Config.PaginationStrategy())

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		pluginID, err := client.UIPlugins.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("plugin")), nil)
		if err != nil {
			return nil, err
		}
		opt := appversions.ListOptions{}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		listAll := func(ctx context.Context, o appversions.ListOptions) *appversions.VersionIterator {
			return client.ApplicationVersions.ListAll(ctx, pluginID, o)
		}
		return c8ystream.ListCall(rawOutput, opt, listAll), nil
	})
}
