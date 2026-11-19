// v2-based tenants enable: activates a tenant by PUTting status=ACTIVE via
// Tenants.Update. The id flag drives iteration (pipe or --id); when no id is
// given the current tenant is used, matching the spec command's default.
package enable

import (
	"context"

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
)

// EnableCmd command
type EnableCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewEnableCmd creates a command to Enable/Activate tenant
func NewEnableCmd(f *cmdutil.Factory) *EnableCmd {
	ccmd := &EnableCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable/Activate tenant",
		Long:  `Enable/Activate an existing tenant`,
		Example: heredoc.Doc(`
$ c8y tenants enable --id "mycompany"
Enable a tenant (from the management tenant)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Tenant id (accepts pipeline)")
	cmd.Flags().String("status", "", "")

	completion.WithOptions(
		cmd,
		completion.WithTenantID("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", false),
		flags.WithPipelineAliases("id", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("Enable-Tenant"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.tenant+json", ""),
	)

	_ = cmd.Flags().MarkHidden("status")

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

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStaticStringValue("status", "ACTIVE"),
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
		id := in.String("id")
		if id == "" {
			id = n.factory.GetTenant()
		}
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Tenant] {
				return client.Tenants.Update(ctx, id, body)
			})
		}, nil
	})
}
