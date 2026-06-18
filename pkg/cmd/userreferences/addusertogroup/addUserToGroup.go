// v2-based add-user-to-group: the user flag drives iteration (pipe a user
// collection or pass --user), and each user is added to the user group via
// UserGroups.Users.AssignUser. The membership body is { user: { self } }; the
// user's self link is taken from a piped user's self/id or built from the user
// name via Users.UserSelfLink. A group name reference is resolved to its id via
// the groupByName endpoint (under ResolveContext so it works under --dry).
package addusertogroup

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	groupusers "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/usergroups/users"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
	"github.com/tidwall/sjson"
)

// AddUserToGroupCmd command
type AddUserToGroupCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewAddUserToGroupCmd creates a command to Add user to group
func NewAddUserToGroupCmd(f *cmdutil.Factory) *AddUserToGroupCmd {
	ccmd := &AddUserToGroupCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "addUserToGroup",
		Short: "Add user to group",
		Long:  `Add an existing user to a group`,
		Example: heredoc.Doc(`
$ c8y userreferences addUserToGroup --group 1 --user peterpi@example.com
Add a user to a user group

$ c8y users list | c8y userreferences addUserToGroup --group admins
Add a list of users to admins group (using pipeline)

$ c8y users list | c8y userreferences addUserToGroup --group business | c8y userreferences addUserToGroup --group admins
Add a list of users to business and admins group (using pipeline)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("group", "", "Group ID (required)")
	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().StringSlice("user", []string{""}, "User id (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithUserGroup("group", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUser("user", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("user", "user.self", true, "user.id", "id", "self"),
		flags.WithPipelineAliases("group", "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPipelineAliases("user", "user.id", "id", "self"),
		flags.WithPowershellName("Add-UserToGroup"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.userReference+json", ""),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("group")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *AddUserToGroupCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	// The user flag drives iteration: a piped user collection feeds it (self / id
	// extracted), or its own --user values drive the run.
	if err := r.InputFlag("user"); err != nil {
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
		groupID, err := client.UserGroups.ResolveID(in.ResolveContext(), tenant, c8ystream.NameOrID(in.String("group")))
		if err != nil {
			return nil, err
		}
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		// Carry the user's self link on the reference body. A piped user supplies
		// its self/id; a user name is turned into its canonical self link.
		if userRef := in.String("user"); userRef != "" {
			if body, err = sjson.SetBytes(body, "user.self", client.Users.UserSelfLink(tenant, userRef)); err != nil {
				return nil, err
			}
		}
		opt := groupusers.AssignUserOptions{
			TenantID: tenant,
			GroupID:  groupID,
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.UserReference] {
				return client.UserGroups.Users.AssignUser(ctx, opt, body)
			})
		}, nil
	})
}
