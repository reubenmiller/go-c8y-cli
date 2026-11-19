// v2-based device profile create: builds the body (name + deviceType filter,
// plus the c8y_Profile type and the c8y_DeviceProfile / c8y_Filter fragments)
// and creates the managed object via DeviceProfiles.Create. The name flag drives
// iteration (pipe or --name) and is a body target.
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

// NewCreateCmd creates a command to Create device profile
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create device profile",
		Long:  `Create a new device profile (managedObject)`,
		Example: heredoc.Doc(`
$ c8y deviceprofiles create --name "python3-requests"
Create a device profile
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "name (accepts pipeline)")
	cmd.Flags().String("deviceType", "", "Device type filter. Only allow device profile to be applied to devices of this type")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("name", "name", false, "name"),
		flags.WithPowershellName("New-DeviceProfile"),
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
		flags.WithOverrideValue("name", "name"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("deviceType", "c8y_Filter.type"),
		flags.WithDefaultTemplateString(`
{type: 'c8y_Profile', c8y_DeviceProfile:{}, c8y_Filter:{}}`),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("type", "name"),
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
				return client.DeviceProfiles.Create(ctx, body)
			})
		}, nil
	})
}
