// v2-based feature enable: the key flag drives iteration (pipe or --key); each
// feature is enabled for the current tenant via Features.Update with an
// {"active":true} body (mergeable with --data/--template).
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

// NewEnableCmd creates a command to Enable tenant feature
func NewEnableCmd(f *cmdutil.Factory) *EnableCmd {
	ccmd := &EnableCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable tenant feature",
		Long:  `Enable tenant feature`,
		Example: heredoc.Doc(`
$ c8y features enable --key example
Enable a feature in the current tenant
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("key", "", "Feature ID / Key (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithFeature("key", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("key", "key", true, "id"),
		flags.WithPowershellName("Enable-CurrentTenantFeature"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.tenant+json", ""),
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

	if err := r.InputFlag("key"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithDefaultTemplateString(`
{"active":true}`),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("active"),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		key := in.String("key")
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Feature] {
				return client.Features.Update(ctx, key, body)
			})
		}, nil
	})
}
