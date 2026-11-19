// v2-based tenant options by category: the category flag drives iteration (pipe
// or --category); the options for each category are fetched via
// Tenants.Options.ListByCategory, which returns a flat key/value map.
package getforcategory

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/tenantoptions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetForCategoryCmd command
type GetForCategoryCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetForCategoryCmd creates a command to Get tenant options by category
func NewGetForCategoryCmd(f *cmdutil.Factory) *GetForCategoryCmd {
	ccmd := &GetForCategoryCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "getForCategory",
		Short: "Get tenant options by category",
		Long:  `Get existing tenant options for a category`,
		Example: heredoc.Doc(`
$ c8y tenantoptions getForCategory --category "c8y_cli_tests"
Get a list of options for a category

$ echo -e "c8y_cli_tests\ncategory2" | c8y tenantoptions getForCategory
Get a list of options for a category
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("category", "", "Tenant Option category (required) (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("category", "category", true, "id"),
		flags.WithPowershellName("Get-TenantOptionForCategory"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.optionCollection+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetForCategoryCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("category"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := tenantoptions.ListByCategoryOptions{
			Category: in.String("category"),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromJSONValue(client.Tenants.Options.ListByCategory(ctx, opt), apiv2.IsDryRun(ctx))
		}, nil
	})
}
