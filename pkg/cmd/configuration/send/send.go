// v2-based configuration send: builds a device operation carrying a
// c8y_DownloadConfigFile fragment and creates it via Operations.CreateRaw (which
// resolves the device name -> id). When a configuration (id or name) is given,
// its url and type are looked up from the configuration repository and written
// onto the operation; otherwise the url and configurationType are taken from the
// flags. The device flag is the iterating input, so the same configuration can be
// sent to many devices from a pipe.
package send

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/configuration"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// SendCmd command
type SendCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewSendCmd creates a command to Send configuration to a device via an operation
func NewSendCmd(f *cmdutil.Factory) *SendCmd {
	ccmd := &SendCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send configuration to a device via an operation",
		Long: `Create a new operation to send configuration to an agent or device.

If you provide the reference to the configuration (via id or name), then the configuration's
url and type will be automatically added to the operation.

You may also manually set the url and configurationType rather than looking up the configuration
file in the configuration repository.
`,
		Example: heredoc.Doc(`
$ c8y configuration send --device mydevice --configuration 12345
Send a configuration file to a device

$ c8y devices list | c8y configuration send --configuration 12345
Send a configuration file to multiple devices

$ c8y devices list | c8y configuration send --configuration my-config-name
Send a configuration file (by name) to multiple devices

$ c8y configuration send --device 12345 --configurationType apt-lists --url "http://example.com/myrepo.list"
Send a custom configuration by manually providing the type and url
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("device", "", "Identifies the target device on which this operation should be performed. (accepts pipeline)")
	cmd.Flags().String("description", "", "Text description of the operation.")
	cmd.Flags().String("configurationType", "", "Configuration type. Leave blank to automatically set it if a matching configuration is found in the c8y configuration repository")
	cmd.Flags().String("url", "", "Url to the configuration. Leave blank to automatically set it if a matching configuration is found in the c8y configuration repository")
	cmd.Flags().String("configuration", "", "Configuration name or id")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithConfiguration("configuration", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("device", "deviceId", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Send-Configuration"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.operation+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *SendCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	// The device flag drives iteration: piped device objects feed it (id /
	// source.id extracted), or its own --device value drives a single run — so
	// `c8y devices list | c8y configuration send` sends the same configuration to
	// each device.
	if err := r.InputFlag("device"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("description", "description"),
		flags.WithStringValue("configurationType", "c8y_DownloadConfigFile.type"),
		flags.WithStringValue("url", "c8y_DownloadConfigFile.url"),
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

	resolveDevice := func(ctx context.Context, ref string) (string, error) {
		return client.Operations.DeviceResolver.ResolveID(ctx, managedobjects.DeviceRef(ref), nil)
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		device := in.String("device")
		if device == "" {
			return nil, cmderrors.NewUserError("Body is missing required properties: deviceId")
		}
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		// Carry the (resolved) device id on the operation. ResolveBodyRef applies
		// the CLI name-or-id convention and resolves the name against the real API
		// even under --dry.
		if body, err = sjson.SetBytes(body, "deviceId", device); err != nil {
			return nil, err
		}
		if body, err = in.ResolveBodyRef(body, "deviceId", resolveDevice); err != nil {
			return nil, err
		}

		// Look up the configuration's url and type (and name, for the default
		// description) from the repository when a configuration reference is given,
		// matching v1. Explicit --url / --configurationType are otherwise used as-is.
		var configName string
		if ref := in.String("configuration"); ref != "" {
			res := client.Configuration.Get(in.ResolveContext(), configuration.ConfigurationRef(c8ystream.NameOrID(ref)), configuration.GetOptions{})
			if res.Err != nil {
				return nil, res.Err
			}
			configName = res.Data.Name()
			if t := res.Data.Get("configurationType").String(); t != "" {
				if body, err = sjson.SetBytes(body, "c8y_DownloadConfigFile.type", t); err != nil {
					return nil, err
				}
			}
			if url := res.Data.Get("url").String(); url != "" {
				if body, err = sjson.SetBytes(body, "c8y_DownloadConfigFile.url", url); err != nil {
					return nil, err
				}
			}
		}

		// Default the operation description from the configuration name and type
		// (matching v1's jsonnet template) when not explicitly provided.
		if gjson.GetBytes(body, "description").String() == "" {
			configType := gjson.GetBytes(body, "c8y_DownloadConfigFile.type").String()
			desc := fmt.Sprintf("Send configuration snapshot %s of configuration type %s to device", configName, configType)
			if body, err = sjson.SetBytes(body, "description", desc); err != nil {
				return nil, err
			}
		}

		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Operation] {
				return client.Operations.CreateRaw(ctx, body)
			})
		}, nil
	})
}
