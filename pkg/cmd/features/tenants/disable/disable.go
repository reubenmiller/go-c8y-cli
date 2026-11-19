// v2-based feature-by-tenant disable: the tenant flag drives iteration (pipe or
// --tenant); the feature (--key) is disabled for each tenant via
// Features.Tenants.DisableForTenant (management tenant). Success yields no output.
package disable

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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DisableCmd command
type DisableCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDisableCmd creates a command to Disable a feature override for a given tenant
func NewDisableCmd(f *cmdutil.Factory) *DisableCmd {
	ccmd := &DisableCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable a feature override for a given tenant",
		Long:  `Disable a feature toggle override for a specific tenant (from the management tenant)`,
		Example: heredoc.Doc(`
$ c8y features tenants disable --key example --tenant t12345
Disable a feature for a specific tenant
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("key", "", "Feature ID / Key")
	cmd.Flags().String("tenant", "", "Unique identifier of a Cumulocity tenant (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithFeature("key", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("tenant", "tenant", false, "tenantId", "owner.tenant.id", "tenant", "id"),
		flags.WithPowershellName("Disable-FeatureByTenant"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DisableCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("tenant"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		key := in.String("key")
		tenant := in.String("tenant")
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Features.Tenants.DisableForTenant(ctx, key, tenant)
			})
		}, nil
	})
}
