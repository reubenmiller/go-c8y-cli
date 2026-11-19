// v2-based tenants disableApplication (deprecated alias of `tenants applications
// disable`): unsubscribes an application from a tenant via
// Tenants.UnsubscribeApplication. The application flag drives iteration (pipe or
// --application); the reference (id or name) is resolved to an id and used as the
// path parameter. The tenant defaults to the session tenant.
package disableapplication

import (
	"context"
	"fmt"

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

// DisableApplicationCmd command
type DisableApplicationCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDisableApplicationCmd creates a command to Unsubscribe application
func NewDisableApplicationCmd(f *cmdutil.Factory) *DisableApplicationCmd {
	ccmd := &DisableApplicationCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "disableApplication",
		Short:  "Unsubscribe application",
		Long:   `Disable/unsubscribe an application from a tenant`,
		Hidden: true,

		Example: heredoc.Doc(`
$ c8y tenants disableApplication --tenant "t12345" --application "myMicroservice"
Disable an application of a tenant by name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant id. Defaults to current tenant (based on credentials)")
	cmd.Flags().String("application", "", "Application id (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithApplication("application", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("application", "application", true, "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPipelineAliases("application", "id"),
		flags.WithDeprecationNotice("please use 'c8y tenants applications disable' instead"),
		flags.WithPowershellName("Disable-Application"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DisableApplicationCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("application"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		ref := in.String("application")
		if ref == "" {
			return nil, fmt.Errorf("required flag(s) \"application\" not set")
		}
		appID, err := client.Applications.ResolveID(in.ResolveContext(), c8ystream.NameOrID(ref), nil)
		if err != nil {
			return nil, err
		}
		tenant := in.String("tenant")
		if tenant == "" {
			tenant = n.factory.GetTenant()
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Tenants.UnsubscribeApplication(ctx, tenant, appID)
			})
		}, nil
	})
}
