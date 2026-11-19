// v2-based tenant option create: builds the body (category/key/value +
// --data/--template) and creates the option via Tenants.Options.Create. The key
// flag drives iteration (pipe or --key).
package create

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// CreateCmd command
type CreateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateCmd creates a command to Create tenant option
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create tenant option",
		Long:  ``,
		Example: heredoc.Doc(`
$ c8y tenantoptions create --category "c8y_cli_tests" --key "option1" --value "1"
Create a tenant option
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("category", "", "Category of option")
	cmd.Flags().String("key", "", "Key of option (accepts pipeline)")
	cmd.Flags().String("value", "", "Value of option")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("key", "key", false, "id"),
		flags.WithPowershellName("New-TenantOption"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.option+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("key"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithOverrideValue("key", "key"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("category", "category"),
		flags.WithStringValue("key", "key"),
		flags.WithAnyStringValue("value", "value"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("category", "key"),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.TenantOption] {
				return client.Tenants.Options.Create(ctx, body)
			})
		}, nil
	})
}
