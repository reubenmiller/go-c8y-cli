// v2-based tenant options list: lists all tenant options via
// Tenants.Options.ListAll. The endpoint has no filter parameters, so the typed
// ListOptions carries only pagination; the command runs once.

package tenantoptions

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/tenantoptions"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get tenant option collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get tenant option collection",
		Long:  `Get collection of tenant options`,
		Example: heredoc.Doc(`
$ c8y tenantoptions list
Get a list of tenant options
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithCollectionProperty("options"),
		flags.WithPowershellName("Get-TenantOptionCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.optionCollection+json", "application/vnd.com.nsn.cumulocity.option+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	paging, err := r.Pagination()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := tenantoptions.ListOptions{}
		opt.PaginationOptions = paging
		return func(ctx context.Context) any {
			return client.Tenants.Options.ListAll(ctx, opt)
		}, nil
	})
}
