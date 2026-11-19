// v2-based software create: builds the software package body (name/description/
// softwareType/deviceType + --data/--template, defaulting type c8y_Software) and
// creates the managed object via Repository.Software.Create. The deviceType flag
// is the iterating input, so the same package can be created for several device
// types from a pipe.
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

// NewCreateCmd creates a command to Create software package
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create software package",
		Long:  `Create a new software package (managedObject)`,
		Example: heredoc.Doc(`
$ c8y software create --name "python3-requests" --description "python requests library" --softwareType apt
Create a software package

$ c8y software create --name "python3-requests" --description "python requests library" --deviceType "c8y_Linux" --softwareType apt
Create a software package which is only applicable for a specific device type

$ echo -e "c8y_Linux\nc8y_MacOS" | c8y software create --name "python3-requests" --description "python requests library"  --softwareType rpm
Create the same software package for multiple device types

$ c8y software create --name "python3-requests" | c8y software versions create --version "1.0.0" --file "python3-requests.deb"
Create a software package and create a new version
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "name")
	cmd.Flags().String("description", "", "Description of the software package")
	cmd.Flags().String("softwareType", "", "Software type")
	cmd.Flags().String("deviceType", "", "Device type filter. Only allow software to be applied to devices of this type (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("deviceType", "c8y_Filter.type", false, "c8y_Filter.type", "deviceType", "type"),
		flags.WithPowershellName("New-Software"),
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
		flags.WithStringValue("softwareType", "softwareType"),
		flags.WithStringValue("deviceType", "c8y_Filter.type"),
		flags.WithDefaultTemplateString(`
{type: 'c8y_Software', c8y_Global:{}}`),
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
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Software] {
				return client.Repository.Software.Create(ctx, body)
			})
		}, nil
	})
}
