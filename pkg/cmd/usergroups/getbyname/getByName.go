// v2-based user group get-by-name: the name flag drives iteration (pipe or
// --name); each name is looked up via the dedicated groupByName endpoint
// (UserGroups.GetByName).
package getbyname

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/usergroups"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetByNameCmd command
type GetByNameCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetByNameCmd creates a command to Get user group by name
func NewGetByNameCmd(f *cmdutil.Factory) *GetByNameCmd {
	ccmd := &GetByNameCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "getByName",
		Short: "Get user group by name",
		Long:  `Get an existing user group by name`,
		Example: heredoc.Doc(`
$ c8y usergroups getByName --name customGroup1
Get user group by its name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().String("name", "", "Group name (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("name", "name", false, "name"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("Get-UserGroupByName"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.group+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetByNameCmd) RunE(cmd *cobra.Command, args []string) error {
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
		opt := usergroups.GetByNameOptions{
			Tenant:    tenant,
			GroupName: in.String("name"),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.UserGroups.GetByName(ctx, opt))
		}, nil
	})
}
