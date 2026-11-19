// v2-based UI plugin update: the id flag drives iteration (pipe or --id) and the
// body builder (name/key/availability/contextPath + --data/--template) is
// evaluated per item, then both feed UIPlugins.Update (the id/name reference is
// resolved internally by the SDK).
package update

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

// UpdateCmd command
type UpdateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUpdateCmd creates a command to Update UI plugin details
func NewUpdateCmd(f *cmdutil.Factory) *UpdateCmd {
	ccmd := &UpdateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update UI plugin details",
		Long:  `Update details of an existing UI plugin`,
		Example: heredoc.Doc(`
$ c8y ui plugins update --id "my-example-app" --availability SHARED
Update plugin availability to SHARED
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Plugin (required) (accepts pipeline)")
	cmd.Flags().String("name", "", "Name of the plugin")
	cmd.Flags().String("key", "", "Shared secret of the plugin")
	cmd.Flags().String("availability", "", "Access level for other tenants")
	cmd.Flags().String("contextPath", "", "contextPath of the plugin")

	completion.WithOptions(
		cmd,
		completion.WithUIPlugin("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("availability", "SHARED", "PRIVATE", "MARKET"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPipelineAliases("id", "id"),
		flags.WithPowershellName("Update-UIPlugin"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.application+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("key", "key"),
		flags.WithStringValue("availability", "availability"),
		flags.WithStringValue("contextPath", "contextPath"),
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
		ref := c8ystream.NameOrID(in.String("id"))
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.UIPlugin] {
				return client.UIPlugins.Update(ctx, ref, body)
			})
		}, nil
	})
}
