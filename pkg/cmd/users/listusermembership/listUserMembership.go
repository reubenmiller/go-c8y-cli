// v2-based user membership list: the id flag drives iteration (pipe or --id);
// the groups a user belongs to are listed via Users.ListGroupsWithUserAll. For
// Cumulocity users the id is the username, passed straight through.
package listusermembership

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

// ListUserMembershipCmd command
type ListUserMembershipCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListUserMembershipCmd creates a command to get user membership collection
func NewListUserMembershipCmd(f *cmdutil.Factory) *ListUserMembershipCmd {
	ccmd := &ListUserMembershipCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "listUserMembership",
		Short: "get user membership collection",
		Long:  `Get information about all groups that a user is a member of`,
		Example: heredoc.Doc(`
$ c8y users listUserMembership --id "myuser"
Get a list of groups that a user belongs to
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "User (required) (accepts pipeline)")
	cmd.Flags().String("tenant", "", "Tenant")

	completion.WithOptions(
		cmd,
		completion.WithUser("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithCollectionProperty("references.#.group"),
		flags.WithPowershellName("Get-UserMembershipCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.groupReferenceCollection+json", "application/vnd.com.nsn.cumulocity.groupReference+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListUserMembershipCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
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
		opt := users.ListGroupsOptions{
			Tenant: tenant,
			UserID: in.String("id"),
		}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.Users.ListGroupsWithUserAll), nil
	})
}
