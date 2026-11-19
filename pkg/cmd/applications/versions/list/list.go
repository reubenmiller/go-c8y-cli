// v2-based application version list: the application flag drives iteration (pipe
// or --application); each application reference (id or name, resolved via
// Applications.ResolveID) drives a typed ApplicationVersions.ListAll call whose
// pages stream through the shared output pipeline.
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

// NewListCmd creates a command to Get application version collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get application version collection",
		Long:  `Get a collection of application versions by a given filter`,
		Example: heredoc.Doc(`
$ c8y applications versions list --application 1234 --pageSize 100
Get application versions

$ c8y applications versions list --application cockpit
Get an application versions by name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("application", "", "Application (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithApplicationWithVersions("application", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("application", "application", false, "id", "name"),
		flags.WithCollectionProperty("applicationVersions"),
		flags.WithPowershellName("Get-ApplicationVersionCollection"),
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

	if err := r.InputFlag("application"); err != nil {
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
		appID, err := client.Applications.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("application")), nil)
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
			return client.ApplicationVersions.ListAll(ctx, appID, o)
		}
		return c8ystream.ListCall(rawOutput, opt, listAll), nil
	})
}
