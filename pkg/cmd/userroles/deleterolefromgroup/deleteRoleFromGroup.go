// v2-based delete-role-from-group: the role flag drives iteration (pipe role
// names/ids or pass --role), and each role is unassigned from the user group via
// UserRoles.Groups.UnassignRole. A group name reference is resolved to its id via
// the groupByName endpoint (under ResolveContext so it works under --dry); the
// role id is the role name (used as-is in the path). Success yields no output.
package deleterolefromgroup

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	roleusergroups "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/userroles/usergroups"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteRoleFromGroupCmd command
type DeleteRoleFromGroupCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteRoleFromGroupCmd creates a command to Unassign role from group
func NewDeleteRoleFromGroupCmd(f *cmdutil.Factory) *DeleteRoleFromGroupCmd {
	ccmd := &DeleteRoleFromGroupCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "deleteRoleFromGroup",
		Short: "Unassign role from group",
		Long:  `Unassign/delete role from a group`,
		Example: heredoc.Doc(`
$ c8y userroles deleteRoleFromGroup --group 12345 --role "ROLE_MEASUREMENT_READ"
Remove a role from the given user group
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("role", []string{""}, "Role name, e.g. ROLE_TENANT_MANAGEMENT_ADMIN (required) (accepts pipeline)")
	cmd.Flags().String("group", "", "Group id (required)")
	cmd.Flags().String("tenant", "", "Tenant")

	completion.WithOptions(
		cmd,
		completion.WithUserGroup("group", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUserRole("role", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("role", "role", true, "id"),
		flags.WithPipelineAliases("group", "id"),
		flags.WithPipelineAliases("role", "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("Remove-RoleFromGroup"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("group")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteRoleFromGroupCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("role"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		tenant := in.String("tenant")
		if tenant == "" {
			tenant = n.factory.GetTenant()
		}
		groupID, err := client.UserGroups.ResolveID(in.ResolveContext(), tenant, c8ystream.NameOrID(in.String("group")))
		if err != nil {
			return nil, err
		}
		opt := roleusergroups.UnassignRoleOptions{
			TenantID: tenant,
			GroupID:  groupID,
			RoleID:   client.UserRoles.RoleID(in.String("role")),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.UserRoles.Groups.UnassignRole(ctx, opt)
			})
		}, nil
	})
}
