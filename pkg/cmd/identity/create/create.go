// v2-based identity create: the device flag drives iteration (pipe or --device).
// The CLI builds the external-identity body (type + externalId, via flags,
// --data or --template), the device reference is resolved (name -> id), and the
// identity is created against that managed object.
package create

import (
	"context"
	"encoding/json"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/identity"
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

// NewCreateCmd creates a command to Create external identity
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create external identity",
		Long:  `Create a new external id for an existing managed object`,
		Example: heredoc.Doc(`
$ c8y identity create --device 1234 --type test --name myserialnumber
Create external identity

$ c8y devices list | c8y identity create --type c8y_Serial --template "{ externalId: input.value.name }"
Create an external identity by using the .name property of the device (via the input template variable)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "The ManagedObject linked to the external ID. (required) (accepts pipeline)")
	cmd.Flags().String("type", "", "The type of the external identifier as string, e.g. 'com_cumulocity_model_idtype_SerialNumber'.")
	cmd.Flags().String("name", "", "The identifier used in the external system that Cumulocity interfaces with.")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("device", "device", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("New-ExternalId"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.externalid+json", ""),
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

	if err := r.InputFlag("device"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("type", "type"),
		flags.WithStringValue("name", "externalId"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithDefaultTemplateString(`
{type: 'c8y_Serial'}`),
		flags.WithRequiredProperties("type", "externalId"),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		deviceID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("device")), nil)
		if err != nil {
			return nil, err
		}
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		var opts identity.IdentityOptions
		if err := json.Unmarshal(body, &opts); err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Identity] {
				return client.Identity.Create(ctx, deviceID, opts)
			})
		}, nil
	})
}
