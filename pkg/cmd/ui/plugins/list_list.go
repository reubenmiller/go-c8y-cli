// v2-based UI plugin list: fills the typed plugins.ListOptions (name/owner/
// availability/... filters) and streams UIPlugins.ListAll. UI plugins are HOSTED
// applications with versions, so the type/hasVersions filters are pinned by the
// SDK. The name flag is the iterating input (pipe or --name), so the command runs
// once for the common no-pipe case.

package plugins

import (
	"context"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/ui/plugins"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get UI plugin collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get UI plugin collection",
		Long:  `Get a collection of UI plugins by a given filter`,
		Example: heredoc.Doc(`
$ c8y ui plugins list --pageSize 100
Get UI plugins

$ c8y ui plugins list --availability SHARED
Get shared ui plugins

$ c8y ui plugins list --availability PRIVATE
Get private ui plugins
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("type", "", "Application type")
	cmd.Flags().String("name", "", "The name of the plugin. (accepts pipeline)")
	cmd.Flags().String("owner", "", "The ID of the tenant that owns the plugin.")
	cmd.Flags().String("providedFor", "", "The ID of a tenant that is subscribed to the plugin but doesn't own them.")
	cmd.Flags().String("subscriber", "", "The ID of a tenant that is subscribed to the plugin.")
	cmd.Flags().StringSlice("user", []string{""}, "The ID of a user that has access to the plugin.")
	cmd.Flags().String("tenant", "", "The ID of a tenant that either owns the plugin or is subscribed to the plugins.")
	cmd.Flags().Bool("hasVersions", true, "When set to true, the returned result contains plugins with an applicationVersions field that is not empty. When set to false, the result will contain applications with an empty applicationVersions field.")
	cmd.Flags().String("availability", "", "Plugin access level for other tenants.")

	completion.WithOptions(
		cmd,
		completion.WithApplication("name", ccmd.factory.ClientV2),
		completion.WithTenantID("owner", ccmd.factory.ClientV2),
		completion.WithTenantID("providedFor", ccmd.factory.ClientV2),
		completion.WithTenantID("subscriber", ccmd.factory.ClientV2),
		completion.WithUser("user", ccmd.factory.ClientV2),
		completion.WithTenantID("tenant", ccmd.factory.ClientV2),
		completion.WithValidateSet("availability", "SHARED", "PRIVATE", "MARKET"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("name", "name", false, "id"),
		flags.WithCollectionProperty("applications"),
		flags.WithPowershellName("Get-UIPluginCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.applicationCollection+json", "application/vnd.com.nsn.cumulocity.application+json"),
	)

	_ = cmd.Flags().MarkHidden("type")
	_ = cmd.Flags().MarkHidden("hasVersions")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("name"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	paging, err := r.Pagination()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := plugins.ListOptions{
			// UI plugins are HOSTED applications that carry versions; the SDK List
			// pins hasVersions=true, and the type filter is pinned here to match v1.
			Type:         plugins.ApplicationTypeHosted,
			Name:         in.String("name"),
			Owner:        in.String("owner"),
			ProvidedFor:  in.String("providedFor"),
			Subscriber:   in.String("subscriber"),
			User:         strings.Join(in.StringSlice("user"), ","),
			Tenant:       in.String("tenant"),
			Availability: in.String("availability"),
		}
		opt.PaginationOptions = paging
		return func(ctx context.Context) any {
			return client.UIPlugins.ListAll(ctx, opt)
		}, nil
	})
}
