// v2-based tenants list: fills the typed tenants.ListOptions (company/domain/
// parent filters). The parent filter defaults to the current tenant, matching
// the spec command. No pipeline input, so the command runs once.
package list

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get tenant collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get tenant collection",
		Long:  `Get collection of tenants`,
		Example: heredoc.Doc(`
$ c8y tenants list
Get a list of tenants
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Company name associated with the Cumulocity tenant")
	cmd.Flags().String("domain", "", "Domain name of the Cumulocity tenant")
	cmd.Flags().String("parent", "", "Identifier of the Cumulocity tenant's parent")

	flags.WithOptions(
		cmd,
		flags.WithCollectionProperty("tenants"),
		flags.WithPowershellName("Get-TenantCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.tenantCollection+json", "application/vnd.com.nsn.cumulocity.tenant+json"),
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
		// The parent filter defaults to the current tenant (spec parity).
		parent := in.String("parent")
		if parent == "" {
			parent = n.factory.GetTenant()
		}
		opt := tenants.ListOptions{
			Company: in.String("name"),
			Domain:  in.String("domain"),
			Parent:  parent,
		}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.Tenants.ListAll), nil
	})
}
