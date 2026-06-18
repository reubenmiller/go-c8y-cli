// v2-based microservice disable: unsubscribes a microservice from a tenant via
// Microservices.Unsubscribe. The id flag drives iteration (pipe or --id); the
// reference (id or name) is resolved internally by the SDK and the tenant
// defaults to the session tenant.
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

// NewDisableCmd creates a command to unsubscribe microservice
func NewDisableCmd(f *cmdutil.Factory) *DisableCmd {
	ccmd := &DisableCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "unsubscribe microservice",
		Long: `Disable (unsubscribe) a microservice from the current tenant
`,
		Example: heredoc.Doc(`
$ c8y microservices disable --id 12345
Disable (unsubscribe) to a microservice

$ c8y microservices disable --id report-agent
Disable (unsubscribe) to a microservice
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Microservice id (required) (accepts pipeline)")
	cmd.Flags().String("tenant", "", "Tenant id")

	completion.WithOptions(
		cmd,
		completion.WithMicroservice("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("id", "id", true, "application.id", "id"),
		flags.WithPipelineAliases("id", "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("Disable-Microservice"),
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

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		ref := c8ystream.NameOrID(in.String("id"))
		tenant := in.String("tenant")
		if tenant == "" {
			tenant = n.factory.GetTenant()
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Microservices.Unsubscribe(ctx, tenant, ref)
			})
		}, nil
	})
}
