// v2-based remote access VNC configuration create: the device flag drives
// iteration (pipe or --device); each device reference is resolved (name -> id)
// and a VNC configuration (built from the typed flags plus the conditional
// credentialsType default — NONE without a password, PASS_ONLY with one) is
// created via RemoteAccess.Configurations.Create.
package create_vnc

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	racfg "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/remoteaccess/remoteaccess_configurations"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// CreateVncCmd command
type CreateVncCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateVncCmd creates a command to Create vnc configuration
func NewCreateVncCmd(f *cmdutil.Factory) *CreateVncCmd {
	ccmd := &CreateVncCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create-vnc",
		Short: "Create vnc configuration",
		Long: `Create a new VNC configuration. If no arguments are provided
then sensible defaults will be used.
`,
		Example: heredoc.Doc(`
$ c8y remoteaccess configurations create-vnc --device device01
Create a VNC configuration that does not require a password

$ c8y remoteaccess configurations create-vnc --device device01 --password 'asd08dcj23dsf{@#9}'
Create a VNC configuration that requires a password
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device (accepts pipeline)")
	cmd.Flags().String("name", "webvnc", "Connection name")
	cmd.Flags().String("hostname", "127.0.0.1", "Hostname")
	cmd.Flags().Int("port", 5900, "Port")
	cmd.Flags().String("password", "", "VNC Password")
	cmd.Flags().String("protocol", "VNC", "Protocol")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("protocol", "VNC"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("device", "device", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("New-RemoteAccessVNCConfiguration"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateVncCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("device"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("hostname", "hostname"),
		flags.WithIntValue("port", "port"),
		flags.WithStringValue("password", "password"),
		flags.WithStringValue("protocol", "protocol"),
		flags.WithDefaultTemplateString(`
{credentialsType: if std.isEmpty(std.get($, 'password', '')) then 'NONE' else 'PASS_ONLY'}`),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("name", "hostname", "port", "protocol", "credentialsType"),
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
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.RemoteAccessConfiguration] {
				return client.RemoteAccess.Configurations.Create(ctx, racfg.CreateOptions{
					ManagedObjectID: deviceID,
					Body:            body,
				})
			})
		}, nil
	})
}
