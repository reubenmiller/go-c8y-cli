// v2-based smart group create: builds the body (name + query + invisible, plus
// the c8y_DynamicGroup type / c8y_IsDynamicGroup fragment) and creates the
// managed object via SmartGroups.Create. The query flag drives iteration (pipe
// or --query).
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

// NewCreateCmd creates a command to Create smart group
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create smart group",
		Long:  `Create a smart group (managed object) which groups devices by an inventory query.`,
		Example: heredoc.Doc(`
$ c8y smartgroups create --name mygroup --query "has(c8y_IsDevice)"
Create a smart group
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Smart group name")
	cmd.Flags().String("query", "", "Smart group query. Should be a valid inventory query. i.e. \"name eq 'myname' and has(myFragment)\" (accepts pipeline)")
	cmd.Flags().Bool("invisible", false, "Should the smart group be hidden from the user interface")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("query", "c8y_DeviceQueryString", false, "id"),
		flags.WithPowershellName("New-SmartGroup"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.inventory+json", ""),
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

	if err := r.Input(); err != nil {
		return err
	}

	err = r.Body(
		flags.WithOverrideValue("query", "c8y_DeviceQueryString"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("query", "c8y_DeviceQueryString"),
		flags.WithBoolValue("invisible", "c8y_IsDynamicGroup.invisible", "{}"),
		flags.WithRequiredTemplateString(`
{type: 'c8y_DynamicGroup', c8y_DeviceQueryString: '', c8y_IsDynamicGroup: {}}`),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("name", "c8y_DeviceQueryString"),
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
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.ManagedObject] {
				return client.SmartGroups.Create(ctx, body)
			})
		}, nil
	})
}
