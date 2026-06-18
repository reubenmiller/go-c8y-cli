// v2-based add-role-to-user: the role flag drives iteration (pipe a role
// collection or pass --role), and each role is assigned to the user via
// UserRoles.Users.AssignRole. The assignment body is { role: { self } }; the
// role's self link is taken from a piped role's self/id or built from the role
// name via UserRoles.RoleSelfLink. The user id is the username (passed through).
package addroletouser

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	roleusers "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/userroles/users"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
	"github.com/tidwall/sjson"
)

// AddRoleToUserCmd command
type AddRoleToUserCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewAddRoleToUserCmd creates a command to Add Role to user
func NewAddRoleToUserCmd(f *cmdutil.Factory) *AddRoleToUserCmd {
	ccmd := &AddRoleToUserCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "addRoleToUser",
		Short: "Add Role to user",
		Long:  `Add a role to an existing user`,
		Example: heredoc.Doc(`
$ c8y userroles addRoleToUser --user "peterpi@example.com" --role "ROLE_ALARM_READ"
Add a role (ROLE_ALARM_READ) to a user
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().String("user", "", "User prefix or full username (required)")
	cmd.Flags().StringSlice("role", []string{""}, "User role id (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUser("user", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUserRole("role", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("role", "role.self", false, "self", "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPipelineAliases("role", "self", "id"),
		flags.WithPowershellName("Add-RoleToUser"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.roleReference+json", ""),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("user")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *AddRoleToUserCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	// The role flag drives iteration: a piped role collection feeds it (self / id
	// extracted), or its own --role values drive the run.
	if err := r.InputFlag("role"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
	)
	if err != nil {
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
		user := in.String("user")
		if user == "" {
			return nil, cmderrors.NewUserError("Body is missing required properties: user")
		}
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		// Carry the role's self link on the reference body. A piped role supplies
		// its self/id; a role name is turned into its canonical self link.
		if roleRef := in.String("role"); roleRef != "" {
			if body, err = sjson.SetBytes(body, "role.self", client.UserRoles.RoleSelfLink(roleRef)); err != nil {
				return nil, err
			}
		}
		opt := roleusers.AssignRoleOptions{
			TenantID: tenant,
			UserID:   user,
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.RoleReference] {
				return client.UserRoles.Users.AssignRole(apiv2.WithTenant(ctx, tenant), opt, body)
			})
		}, nil
	})
}
