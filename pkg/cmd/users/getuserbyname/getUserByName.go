// v2-based user get-by-name: the name flag drives iteration (pipe or --name);
// each username is looked up via the dedicated userByName endpoint
// (Users.GetByUsername).
package getuserbyname

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/users"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetUserByNameCmd command
type GetUserByNameCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetUserByNameCmd creates a command to Get user by name
func NewGetUserByNameCmd(f *cmdutil.Factory) *GetUserByNameCmd {
	ccmd := &GetUserByNameCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "getUserByName",
		Short: "Get user by name",
		Long:  `Get the user details by referencing their username instead of id`,
		Example: heredoc.Doc(`
$ c8y users getUserByName --name "myuser"
Get a user by name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().String("name", "", "Username (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("name", "name", true, "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("Get-UserByName"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.user+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetUserByNameCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("name"); err != nil {
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
		opt := users.GetByUsernameOptions{
			Tenant:   tenant,
			Username: in.String("name"),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.Users.GetByUsername(ctx, opt))
		}, nil
	})
}
