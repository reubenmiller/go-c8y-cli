// v2-based delete-role-from-user: the role flag drives iteration (pipe role
// names/ids or pass --role), and each role is unassigned from the user via
// UserRoles.Users.UnassignRole. The user id is the username (passed through) and
// the role id is the role name (used as-is in the path). Success yields no output.
package deleterolefromuser

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
	roleusers "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/userroles/users"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteRoleFromUserCmd command
type DeleteRoleFromUserCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteRoleFromUserCmd creates a command to Unassign role from user
func NewDeleteRoleFromUserCmd(f *cmdutil.Factory) *DeleteRoleFromUserCmd {
	ccmd := &DeleteRoleFromUserCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "deleteRoleFromUser",
		Short: "Unassign role from user",
		Long:  `Unassign/delete role from a user`,
		Example: heredoc.Doc(`
$ c8y userroles deleteRoleFromUser --user "peterpi@example.com" --role "ROLE_MEASUREMENT_READ"
Remove a role from the given user
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("user", "", "User (required)")
	cmd.Flags().StringSlice("role", []string{""}, "Role name (required) (accepts pipeline)")
	cmd.Flags().String("tenant", "", "Tenant")

	completion.WithOptions(
		cmd,
		completion.WithUser("user", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUserRole("role", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("role", "role", true, "id"),
		flags.WithPipelineAliases("role", "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("Remove-RoleFromUser"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("user")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteRoleFromUserCmd) RunE(cmd *cobra.Command, args []string) error {
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
		opt := roleusers.UnassignRoleOptions{
			TenantID: tenant,
			UserID:   in.String("user"),
			RoleID:   client.UserRoles.RoleID(in.String("role")),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.UserRoles.Users.UnassignRole(ctx, opt)
			})
		}, nil
	})
}
