// v2-based tenant application-reference list: the tenant flag drives iteration
// (pipe or --tenant); each tenant's referenced applications are listed via
// Tenants.ListAllApplicationReferences. When no tenant is given the current
// tenant is used. Hidden + deprecated in favour of `c8y tenants applications
// list`; kept for back-compat.
package listreferences

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/spf13/cobra"
)

// ListReferencesCmd command
type ListReferencesCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListReferencesCmd creates a command to Get application reference collection
func NewListReferencesCmd(f *cmdutil.Factory) *ListReferencesCmd {
	ccmd := &ListReferencesCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "listReferences",
		Short:  "Get application reference collection",
		Long:   `Get a collection of application references on a tenant`,
		Hidden: true,
		Example: heredoc.Doc(`
$ c8y tenants listReferences --tenant "mycompany"
Get a list of referenced applications on a given tenant (from management tenant)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant id (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("tenant", "tenant", false, "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithCollectionProperty("references"),
		flags.WithDeprecationNotice("please use 'c8y tenants applications list' instead"),
		flags.WithPowershellName("Get-ApplicationReferenceCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.applicationReferenceCollection+json", "application/vnd.com.nsn.cumulocity.applicationReference+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListReferencesCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("tenant"); err != nil {
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
		tenant := in.String("tenant")
		if tenant == "" {
			tenant = n.factory.GetTenant()
		}
		opt := tenants.ListApplicationReferencesOptions{}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		all := func(ctx context.Context, o tenants.ListApplicationReferencesOptions) *pagination.Iterator[jsonmodels.ApplicationReference] {
			return client.Tenants.ListAllApplicationReferences(ctx, tenant, o)
		}
		return c8ystream.ListCall(rawOutput, opt, all), nil
	})
}
