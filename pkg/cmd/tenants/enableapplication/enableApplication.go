// v2-based tenants enableApplication (deprecated alias of `tenants applications
// enable`): subscribes an application to a tenant via Tenants.SubscribeApplication.
// The application flag drives iteration (pipe or --application); the reference is
// resolved to an id and written into the body's application.id (so piped input
// populates it via the driver), merged with any --data/--template body, then
// POSTed to the tenant's applications. The tenant defaults to the session tenant.
package enableapplication

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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
	"github.com/tidwall/sjson"
)

// EnableApplicationCmd command
type EnableApplicationCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewEnableApplicationCmd creates a command to Subscribe application
func NewEnableApplicationCmd(f *cmdutil.Factory) *EnableApplicationCmd {
	ccmd := &EnableApplicationCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "enableApplication",
		Short:  "Subscribe application",
		Long:   `Enable/subscribe an application to a tenant`,
		Hidden: true,

		Example: heredoc.Doc(`
$ c8y tenants enableApplication --tenant "t12345" --application "myMicroservice"
Enable an application of a tenant by name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
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
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("application", "application.id", true, "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPipelineAliases("application", "id"),
		flags.WithDeprecationNotice("please use 'c8y tenants applications enable' instead"),
		flags.WithPowershellName("Enable-Application"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.applicationReference+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *EnableApplicationCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("application"); err != nil {
		return err
	}

	// --data/--template build any extra fields; application.id is set in Build
	// from the resolved application reference.
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
		ref := in.String("application")
		if ref == "" {
			return nil, fmt.Errorf("Body is missing required properties: application.id")
		}
		appID, err := client.Applications.ResolveID(in.ResolveContext(), c8ystream.NameOrID(ref), nil)
		if err != nil {
			return nil, err
		}
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		body, err = sjson.SetBytes(body, "application.id", appID)
		if err != nil {
			return nil, err
		}
		tenant := in.String("tenant")
		if tenant == "" {
			tenant = n.factory.GetTenant()
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.ApplicationReference] {
				return client.Tenants.SubscribeApplication(ctx, tenant, body)
			})
		}, nil
	})
}
