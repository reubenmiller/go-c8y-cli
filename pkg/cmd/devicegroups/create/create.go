// v2-based device group create: builds the body (name + type, plus the default
// c8y_DeviceGroup type / c8y_IsDeviceGroup fragment template) and creates the
// managed object via DeviceGroups.Create. The name flag drives iteration (pipe
// or --name).
package create

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
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

// NewCreateCmd creates a command to Create device group
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create device group",
		Long:  `Create a new device group to logically group one or more devices`,
		Example: heredoc.Doc(`
$ c8y devicegroups create --name mygroup
Create device group

$ c8y devicegroups create --name mygroup --data "custom_value1=1234"
Create device group with custom properties
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Device group name (accepts pipeline)")
	cmd.Flags().String("type", "", "Device group type (c8y_DeviceGroup (root folder) or c8y_DeviceSubGroup (sub folder)). Defaults to c8y_DeviceGroup")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("type", "c8y_DeviceGroup", "c8y_DeviceSubGroup"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("name", "name", false, "name"),
		flags.WithPowershellName("New-DeviceGroup"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.customDeviceGroup+json", ""),
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
		flags.WithStringValue("type", "type"),
		flags.WithDefaultTemplateString(`
{type: 'c8y_DeviceGroup', c8y_IsDeviceGroup: {}}`),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("name"),
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
				return client.DeviceGroups.Create(ctx, body)
			})
		}, nil
	})
}
