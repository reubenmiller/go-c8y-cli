// v2-based microservice enable: subscribes a microservice to a tenant via
// Microservices.Enable. The id flag drives iteration (pipe or --id); the
// microservice reference is resolved to an id and written into the
// application.id body field (so piped input populates it via the driver),
// merged with any --data/--template body, then POSTed to the tenant's
// applications. The tenant defaults to the session tenant.
package enable

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
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// EnableCmd command
type EnableCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewEnableCmd creates a command to subscribe to microservice
func NewEnableCmd(f *cmdutil.Factory) *EnableCmd {
	ccmd := &EnableCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "subscribe to microservice",
		Long: `Enabling (subscribing) a microservice will activate the application in the tenant
`,
		Example: heredoc.Doc(`
$ c8y microservices enable --id 12345
Enable (subscribe) to a microservice

$ c8y microservices enable --id report-agent
Enable (subscribe) to a microservice by name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant id")
	cmd.Flags().String("id", "", "Microservice id (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithMicroservice("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "application.id", false, "application.id", "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPipelineAliases("id", "id"),
		flags.WithPowershellName("Enable-Microservice"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.applicationReference+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *EnableCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	// --data/--template build any extra fields; application.id is set in Build
	// from the resolved microservice reference.
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
		// Build the body first so a --template/--data supplied application.id can
		// satisfy the requirement (v1 accepted the id from the body/template, not
		// only the --id flag). The explicit --id still wins when set.
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		ref := in.String("id")
		if ref == "" {
			ref = gjson.GetBytes(body, "application.id").String()
		}
		if ref == "" {
			return nil, fmt.Errorf("Body is missing required properties: application.id")
		}
		// Resolve the microservice reference (id or name) and write it into the
		// body's application.id; the value comes from the driver flag, so this
		// works for piped input too.
		appID, err := client.Microservices.ResolveID(in.ResolveContext(), c8ystream.NameOrID(ref), nil)
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
				return client.Microservices.Enable(ctx, tenant, body)
			})
		}, nil
	})
}
