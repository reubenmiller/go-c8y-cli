// v2-based remote access WebSSH configuration create: the device flag drives
// iteration (pipe or --device); each device reference is resolved (name -> id)
// and an SSH configuration (built from the typed flags) is created via
// RemoteAccess.Configurations.Create.
package create_webssh

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

// CreateWebsshCmd command
type CreateWebsshCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateWebsshCmd creates a command to Create web ssh configuration
func NewCreateWebsshCmd(f *cmdutil.Factory) *CreateWebsshCmd {
	ccmd := &CreateWebsshCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create-webssh",
		Short: "Create web ssh configuration",
		Long: `Create a new WebSSH configuration. If no arguments are provided
then sensible defaults will be used.
`,
		Example: heredoc.Doc(`
$ c8y remoteaccess configurations create-webssh --device device01 --username admin --password "3Xz7cEj%oAmt#dnUMP*N"
Create a webssh configuration (with username/password authentication)

$ c8y remoteaccess configurations create-webssh --device device01 --hostname 127.0.0.1 --port 2222 --username admin --privateKey "xxxx" --publicKey "yyyyy"
Create a webssh configuration with a custom hostname and port (with ssh key authentication)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device (accepts pipeline)")
	cmd.Flags().String("name", "webssh", "Connection name")
	cmd.Flags().String("hostname", "127.0.0.1", "Hostname")
	cmd.Flags().Int("port", 22, "Port")
	cmd.Flags().String("credentialsType", "USER_PASS", "Credentials type")
	cmd.Flags().String("privateKey", "", "Private ssh key")
	cmd.Flags().String("publicKey", "", "Public ssh key")
	cmd.Flags().String("username", "", "Username")
	cmd.Flags().String("password", "", "Username")
	cmd.Flags().String("protocol", "SSH", "Protocol")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("credentialsType", "USER_PASS", "KEY_PAIR", "CERTIFICATE"),
		completion.WithValidateSet("protocol", "SSH"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("device", "device", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("New-RemoteAccessWebSSHConfiguration"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateWebsshCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithStringValue("credentialsType", "credentialsType"),
		flags.WithStringValue("privateKey", "privateKey"),
		flags.WithStringValue("publicKey", "publicKey"),
		flags.WithStringValue("username", "username"),
		flags.WithStringValue("password", "password"),
		flags.WithStringValue("protocol", "protocol"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("hostname", "port", "protocol", "name", "credentialsType"),
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
