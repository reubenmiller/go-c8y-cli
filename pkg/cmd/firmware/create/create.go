// v2-based firmware create: builds the firmware package body (name/description/
// deviceType + --data/--template, defaulting type c8y_Firmware) and creates the
// managed object via Repository.Firmware.CreateRaw. The deviceType flag is the
// iterating input, so the same package can be created for several device types
// from a pipe.
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

// NewCreateCmd creates a command to Create firmware package
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create firmware package",
		Long:  `Create a new firmware package (managedObject)`,
		Example: heredoc.Doc(`
$ c8y firmware create --name "iot-linux" --description "Linux image for IoT devices"
Create a firmware package

$ echo -e "c8y_Linux\nc8y_MacOS" | c8y firmware create --name "iot-linux" --description "Linux image for IoT devices"
Create the same firmware package for multiple device types
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "name")
	cmd.Flags().String("description", "", "Description of the firmware package")
	cmd.Flags().String("deviceType", "", "Device type filter. Only allow firmware to be applied to devices of this type (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("deviceType", "c8y_Filter.type", false, "c8y_Filter.type", "deviceType", "type"),
		flags.WithPowershellName("New-Firmware"),
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
		flags.WithOverrideValue("deviceType", "c8y_Filter.type"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("description", "description"),
		flags.WithStringValue("deviceType", "c8y_Filter.type"),
		flags.WithDefaultTemplateString(`
{type: 'c8y_Firmware', c8y_Global:{}}`),
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
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Firmware] {
				return client.Repository.Firmware.CreateRaw(ctx, body)
			})
		}, nil
	})
}
