// v2-based group role-reference list: the group flag drives iteration (pipe or
// --group), and each group's assigned role references are listed (paginated) via
// UserRoles.Groups.ListAll. A group name reference is resolved to its id via the
// groupByName endpoint (under ResolveContext so it works under --dry). Output is
// the role reference collection (collectionProperty "references"), matching v1.
package getrolereferencecollectionfromgroup

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	roleusergroups "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/userroles/usergroups"
	"github.com/spf13/cobra"
)

// GetRoleReferenceCollectionFromGroupCmd command
type GetRoleReferenceCollectionFromGroupCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetRoleReferenceCollectionFromGroupCmd creates a command to Get role references from user group
func NewGetRoleReferenceCollectionFromGroupCmd(f *cmdutil.Factory) *GetRoleReferenceCollectionFromGroupCmd {
	ccmd := &GetRoleReferenceCollectionFromGroupCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "getRoleReferenceCollectionFromGroup",
		Short: "Get role references from user group",
		Long:  `Get collection of user role references from a group`,
		Example: heredoc.Doc(`
$ c8y userroles getRoleReferenceCollectionFromGroup --group "12345"
Get a list of role references for a user group
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().StringSlice("group", []string{""}, "Group id (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUserGroup("group", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("group", "group", true, "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPipelineAliases("group", "id"),
		flags.WithCollectionProperty("references"),
		flags.WithPowershellName("Get-RoleReferenceCollectionFromGroup"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.roleReferenceCollection+json", "application/vnd.com.nsn.cumulocity.roleReference+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetRoleReferenceCollectionFromGroupCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("group"); err != nil {
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
		groupID, err := client.UserGroups.ResolveID(in.ResolveContext(), tenant, c8ystream.NameOrID(in.String("group")))
		if err != nil {
			return nil, err
		}
		opt := roleusergroups.ListOptions{
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
		return c8ystream.ListCall(rawOutput, opt, client.UserRoles.Groups.ListAll), nil
	})
}
