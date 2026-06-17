// v2-based configuration create: builds the configuration body (name/description/
// configurationType/url/deviceType + --data/--template, defaulting type
// c8y_ConfigurationDump and the c8y_Global fragment) and creates the managed
// object via Configuration.Create. The deviceType flag is the iterating input, so
// the same configuration can be created for several device types from a pipe.
// When --file is given the binary is uploaded and linked as a child addition by
// the SDK (Configuration.Create), so the upload is streamed via SubmitUpload.
package create

import (
	"context"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/configuration"
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

// NewCreateCmd creates a command to Create configuration file
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create configuration file",
		Long:  `Create a new configuration file (managedObject)`,
		Example: heredoc.Doc(`
$ c8y configuration create --name "agent config" --description "Default agent configuration" --configurationType "agentConfig" --url "https://test.com/content/raw/app.json"
Create a configuration package

$ echo -e "c8y_Linux\nc8y_MacOS\nc8y_Windows" | c8y configuration create --name "default-vpn-config" --configurationType "VPN_CONFIG" --file default.vpn
Create multiple configurations using different device type filters (via pipeline)
The stdin will be mapped to the deviceType property. This was you can easily make the same configuration
available for multiple device types

        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "name")
	cmd.Flags().String("description", "", "Description of the configuration package")
	cmd.Flags().String("configurationType", "", "Configuration type")
	cmd.Flags().String("url", "", "URL link to the configuration file")
	cmd.Flags().String("deviceType", "", "Device type filter. Only allow configuration to be applied to devices of this type (accepts pipeline)")
	cmd.Flags().String("file", "", "File to upload")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("deviceType", "deviceType", false, "deviceType", "c8y_Filter.type", "type"),
		flags.WithPowershellName("New-Configuration"),
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
		flags.WithOverrideValue("deviceType", "deviceType"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("description", "description"),
		flags.WithStringValue("configurationType", "configurationType"),
		flags.WithStringValue("url", "url"),
		flags.WithStringValue("deviceType", "deviceType"),
		flags.WithDefaultTemplateString(`
{type: 'c8y_ConfigurationDump', c8y_Global:{}}`),
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
		file := in.String("file")
		return func(ctx context.Context) output.Seq {
			if file != "" {
				return c8ystream.SubmitUpload(ctx, http.MethodPost, "", func(ctx context.Context) op.Result[jsonmodels.ManagedObject] {
					return client.Configuration.Create(ctx, configuration.CreateOptions{
						Body: body,
						File: configuration.UploadFileOptions{FilePath: file},
					})
				})
			}
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.ManagedObject] {
				return client.Configuration.CreateRaw(ctx, body)
			})
		}, nil
	})
}
