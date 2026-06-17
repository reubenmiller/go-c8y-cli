// v2-based configuration update: the id flag drives iteration (pipe or --id) and
// the body builder (newName/description/configurationType/url/deviceType +
// --data/--template) is evaluated per item, then both feed Configuration.Update
// (the id/name reference is resolved scoped to configuration files). When --file
// is given the binary is uploaded and linked as a child addition by the SDK, so
// the upload is streamed via SubmitUpload.
package update

import (
	"context"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/configuration"
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

// NewUpdateCmd creates a command to Update configuration
func NewUpdateCmd(f *cmdutil.Factory) *UpdateCmd {
	ccmd := &UpdateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update configuration",
		Long:  `Update an existing configuration file (managedObject)`,
		Example: heredoc.Doc(`
$ c8y configuration update --id 12345 --newName "my_custom_name" --data "{\"com_my_props\":{},\"value\":1}"
Update a configuration file
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Configuration package (managedObject) id (required) (accepts pipeline)")
	cmd.Flags().String("newName", "", "New configuration name")
	cmd.Flags().String("description", "", "Description of the configuration package")
	cmd.Flags().String("configurationType", "", "Configuration type")
	cmd.Flags().String("url", "", "URL link to the configuration file")
	cmd.Flags().String("deviceType", "", "Device type filter. Only allow configuration to be applied to devices of this type")
	cmd.Flags().String("file", "", "File to be uploaded")

	completion.WithOptions(
		cmd,
		completion.WithConfiguration("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("Update-Configuration"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.inventory+json", ""),
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
		flags.WithStringValue("newName", "name"),
		flags.WithStringValue("description", "description"),
		flags.WithStringValue("configurationType", "configurationType"),
		flags.WithStringValue("url", "url"),
		flags.WithStringValue("deviceType", "deviceType"),
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
		ref := configuration.ConfigurationRef(c8ystream.NameOrID(in.String("id")))
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		file := in.String("file")
		return func(ctx context.Context) output.Seq {
			if file != "" {
				return c8ystream.SubmitUpload(ctx, http.MethodPut, string(ref), func(ctx context.Context) op.Result[jsonmodels.ManagedObject] {
					return client.Configuration.Update(ctx, ref, configuration.UpdateOptions{
						Body: body,
						File: configuration.UploadFileOptions{FilePath: file},
					})
				})
			}
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.ManagedObject] {
				return client.Configuration.UpdateRaw(ctx, ref, body)
			})
		}, nil
	})
}
