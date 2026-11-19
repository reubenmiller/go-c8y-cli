// v2-based user list: lists the users of a tenant via Users.ListAll. The tenant
// flag drives iteration (pipe tenants/objects to list each one's users),
// defaulting to the current tenant when not given, so the command runs once for
// the common case. The username/groups/owner/onlyDevices/withSubusersCount flags
// are applied as collection filters.
package list

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/users"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get user collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get user collection",
		Long:  `Get a collection of users based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y users list
Get a list of users
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().String("username", "", "prefix or full username")
	cmd.Flags().String("groups", "", "numeric group identifiers separated by commas; result will contain only users which belong to at least one of specified groups")
	cmd.Flags().String("owner", "", "exact username")
	cmd.Flags().Bool("onlyDevices", false, "If set to 'true', result will contain only users created during bootstrap process (starting with 'device_'). If flag is absent (or false) the result will not contain 'device_' users.")
	cmd.Flags().Bool("withSubusersCount", false, "if set to 'true', then each of returned users will contain additional field 'subusersCount' - number of direct subusers (users with corresponding 'owner').")

	completion.WithOptions(
		cmd,
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("tenant", "tenant", false, "tenant", "owner.tenant.id"),
		flags.WithCollectionProperty("users"),
		flags.WithPowershellName("Get-UserCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.userCollection+json", "application/vnd.com.nsn.cumulocity.user+json"),
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
		// An empty tenant defaults to the current tenant (spec parity).
		tenant := in.String("tenant")
		if tenant == "" {
			tenant = n.factory.GetTenant()
		}
		opt := users.ListOptions{
			Tenant:            tenant,
			Username:          in.String("username"),
			Owner:             in.String("owner"),
			OnlyDevices:       in.Bool("onlyDevices"),
			WithSubusersCount: in.Bool("withSubusersCount"),
		}
		// groups is a single comma-separated query parameter (matching v1).
		if g := in.String("groups"); g != "" {
			opt.Groups = []string{g}
		}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.Users.ListAll), nil
	})
}
