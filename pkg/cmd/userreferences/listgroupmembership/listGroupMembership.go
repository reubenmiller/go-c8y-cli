// v2-based group membership list: the id flag drives iteration (pipe groups or
// --id), and each group's member users are listed (paginated) via
// UserGroups.Users.ListAll. A group name reference is resolved to its id via the
// groupByName endpoint (under ResolveContext so it works under --dry). Output is
// the member users (the reference collection is unwrapped to references.#.user),
// matching v1.
package listgroupmembership

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	groupusers "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/usergroups/users"
	"github.com/spf13/cobra"
)

// ListGroupMembershipCmd command
type ListGroupMembershipCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListGroupMembershipCmd creates a command to Get users in group
func NewListGroupMembershipCmd(f *cmdutil.Factory) *ListGroupMembershipCmd {
	ccmd := &ListGroupMembershipCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "listGroupMembership",
		Short: "Get users in group",
		Long:  `Get all users in a user group`,
		Example: heredoc.Doc(`
$ c8y userreferences listGroupMembership --id 1
List the users within a user group

$ c8y usergroups list | c8y userreferences listGroupMembership
List users in user groups (using pipeline)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Group ID (required) (accepts pipeline)")
	cmd.Flags().String("tenant", "", "Tenant")

	completion.WithOptions(
		cmd,
		completion.WithUserGroup("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPipelineAliases("id", "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithCollectionProperty("references.#.user"),
		flags.WithPowershellName("Get-UserGroupMembershipCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.userReferenceCollection+json", "application/vnd.com.nsn.cumulocity.user+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListGroupMembershipCmd) RunE(cmd *cobra.Command, args []string) error {
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
		groupID, err := client.UserGroups.ResolveID(in.ResolveContext(), tenant, c8ystream.NameOrID(in.String("id")))
		if err != nil {
			return nil, err
		}
		opt := groupusers.ListOptions{
			TenantID: tenant,
			GroupID:  groupID,
		}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.UserGroups.Users.ListAll), nil
	})
}
