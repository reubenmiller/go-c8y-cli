// v2-based tenant option editable-flag update: the key flag drives iteration
// (pipe or --key); the read-only/editable setting of the option (by category +
// key) is updated via Tenants.Options.UpdateEditableFlag.
package updateedit

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/tenantoptions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// UpdateEditCmd command
type UpdateEditCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUpdateEditCmd creates a command to Update tenant option edit setting
func NewUpdateEditCmd(f *cmdutil.Factory) *UpdateEditCmd {
	ccmd := &UpdateEditCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "updateEdit",
		Short: "Update tenant option edit setting",
		Long: `Update read-only setting of an existing tenant option
Required role:: ROLE_OPTION_MANAGEMENT_ADMIN, Required tenant management Example Request:: Update access.control.allow.origin option.
`,
		Example: heredoc.Doc(`
$ c8y tenantoptions updateEdit --category "c8y_cli_tests" --key "option8" --editable "true"
Update editable property for an existing tenant option
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("category", "", "Tenant Option category (required)")
	cmd.Flags().String("key", "", "Tenant Option key (required) (accepts pipeline)")
	cmd.Flags().String("editable", "", "Whether the tenant option should be editable or not (required)")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("editable", "true", "false"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("key", "key", true, "id"),
		flags.WithPowershellName("Update-TenantOptionEditable"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.option+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateEditCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("key"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := tenantoptions.UpdateEditableFlagOption{
			Category: in.String("category"),
			Key:      in.String("key"),
			Editable: in.String("editable") == "true",
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.TenantOption] {
				return client.Tenants.Options.UpdateEditableFlag(ctx, opt)
			})
		}, nil
	})
}
