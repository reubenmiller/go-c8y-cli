// v2-based device registration register: the id flag drives iteration (pipe or
// --id), so the same registration can be applied to many device ids. The body
// carries the device id (set from the driver), an optional type, and an optional
// device-group reference (group), which is resolved (name -> id) into groupId
// before the create call via Devices.Registration.CreateRaw.
package register

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
	"github.com/tidwall/sjson"
)

// RegisterCmd command
type RegisterCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewRegisterCmd creates a command to Register device with username/password and manual device approval/bootstrapping
func NewRegisterCmd(f *cmdutil.Factory) *RegisterCmd {
	ccmd := &RegisterCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register device with username/password and manual device approval/bootstrapping",
		Long: `Register a new device (request) where the device is using the manual device bootstrapping
process to retrieve its device credentials (username/password).

See Cumulocity docs for more details: https://cumulocity.com/docs/2024/device-integration/rest/
`,
		Example: heredoc.Doc(`
$ c8y deviceregistration register --id "ASDF098SD1J10912UD92JDLCNCU8"
Register a new device

$ c8y deviceregistration register --id "ASDF098SD1J10912UD92JDLCNCU8" --group "My Group"
Register a new device and assign to a group
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Device identifier. Max: 1000 characters. E.g. IMEI (required) (accepts pipeline)")
	cmd.Flags().String("type", "", "Type of the device")
	cmd.Flags().String("group", "", "Group to which the device will be assigned")

	completion.WithOptions(
		cmd,
		completion.WithDeviceGroup("group", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPipelineAliases("group", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Register-Device"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.newDeviceRequest+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *RegisterCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("type", "type"),
		flags.WithStringValue("group", "groupId"),
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

	resolveGroup := func(ctx context.Context, ref string) (string, error) {
		return client.DeviceGroups.ResolveID(ctx, ref, nil)
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		// The id flag is the iterating driver (a string slice), so carry its
		// resolved value onto the body here rather than via a body getter.
		if body, err = sjson.SetBytes(body, "id", in.String("id")); err != nil {
			return nil, err
		}
		// Resolve the device-group reference (name -> id) the body carries.
		if body, err = in.ResolveBodyRef(body, "groupId", resolveGroup); err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.DeviceRequest] {
				return client.Devices.Registration.CreateRaw(ctx, body)
			})
		}, nil
	})
}
