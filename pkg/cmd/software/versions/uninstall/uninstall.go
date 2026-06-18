// v2-based software version uninstall: builds a device operation carrying a
// c8y_SoftwareUpdate entry (action "delete"), resolves the target device
// (name -> id), and creates the operation via Operations.CreateRaw. The device
// flag is the iterating input, so the same software can be uninstalled from many
// devices from a pipe. Unlike install, no repository lookup is performed.
package uninstall

import (
	"context"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// UninstallCmd command
type UninstallCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUninstallCmd creates a command to Uninstall software version on a device
func NewUninstallCmd(f *cmdutil.Factory) *UninstallCmd {
	ccmd := &UninstallCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall software version on a device",
		Long:  `Uninstall software version on a device`,
		Example: heredoc.Doc(`
$ c8y software versions uninstall --device 1234 --software go-c8y-cli --version 1.0.0
Uninstall a software package version
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("device", "", "Device or agent where the software should be installed (accepts pipeline)")
	cmd.Flags().String("software", "", "Software name (required)")
	cmd.Flags().String("version", "", "Software version name or id")
	cmd.Flags().String("softwareType", "", "Software type. Leave blank to automatically set it if a matching software/version is found in the c8y software repository")
	cmd.Flags().String("action", "delete", "Software action")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithSoftware("software", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithSoftwareVersion("version", "software", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("action", "delete"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("device", "deviceId", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Remove-SoftwareVersion"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.operation+json", ""),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("software")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UninstallCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("device"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("software", "c8y_SoftwareUpdate.0.name"),
		flags.WithStringValue("version", "c8y_SoftwareUpdate.0.version"),
		flags.WithStringValue("softwareType", "c8y_SoftwareUpdate.0.softwareType"),
		flags.WithStringValue("action", "c8y_SoftwareUpdate.0.action"),
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
		if body, err = sjson.SetBytes(body, "deviceId", device); err != nil {
			return nil, err
		}
		if body, err = in.ResolveBodyRef(body, "deviceId", resolveDevice); err != nil {
			return nil, err
		}

		if err := requireBodyKeys(body, "deviceId", "c8y_SoftwareUpdate.0.name", "c8y_SoftwareUpdate.0.action"); err != nil {
			return nil, err
		}

		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Operation] {
				return client.Operations.CreateRaw(ctx, body)
			})
		}, nil
	})
}

// requireBodyKeys returns a user error listing any of the given gjson paths that
// are missing or empty in the body (mirrors v1's bodyRequiredKeys check).
func requireBodyKeys(body []byte, keys ...string) error {
	var missing []string
	for _, k := range keys {
		if v := gjson.GetBytes(body, k); !v.Exists() || v.String() == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return cmderrors.NewUserError("Body is missing required properties: " + strings.Join(missing, ", "))
	}
	return nil
}
