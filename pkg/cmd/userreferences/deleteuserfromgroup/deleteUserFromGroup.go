// v2-based delete-user-from-group: the user flag drives iteration (pipe users
// or pass --user), and each user is removed from the user group via
// UserGroups.Users.UnassignUser. A group name reference is resolved to its id
// via the groupByName endpoint (under ResolveContext so it works under --dry);
// the user id is the user name (used as-is in the path). Success yields no output.
package deleteuserfromgroup

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
	groupusers "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/usergroups/users"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteUserFromGroupCmd command
type DeleteUserFromGroupCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteUserFromGroupCmd creates a command to Delete user from group
func NewDeleteUserFromGroupCmd(f *cmdutil.Factory) *DeleteUserFromGroupCmd {
	ccmd := &DeleteUserFromGroupCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "deleteUserFromGroup",
		Short: "Delete user from group",
		Long:  `Delete an existing user from a user group`,
		Example: heredoc.Doc(`
$ c8y userreferences deleteUserFromGroup --group 1 --user peterpi@example.com
Delete a user from a user group

$ c8y users get --id peterpi@example.com | c8y userreferences deleteUserFromGroup --group 1
Delete a user from a user group (using pipeline)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("user", []string{""}, "User id/username (required) (accepts pipeline)")
	cmd.Flags().String("group", "", "Group ID (required)")
	cmd.Flags().String("tenant", "", "Tenant")

	completion.WithOptions(
		cmd,
		completion.WithUserGroup("group", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUser("user", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("user", "user", true, "id"),
		flags.WithPipelineAliases("group", "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("Remove-UserFromGroup"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("group")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteUserFromGroupCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("user"); err != nil {
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
		opt := groupusers.UnassignUserOptions{
			TenantID: tenant,
			GroupID:  groupID,
			UserID:   client.Users.UserID(in.String("user")),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.UserGroups.Users.UnassignUser(ctx, opt)
			})
		}, nil
	})
}
