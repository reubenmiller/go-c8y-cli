// v2-based firmware version install: builds a device operation that carries the
// firmware name/version/url, resolves the target device (name -> id), and creates
// the operation via Operations.CreateRaw. When --url is omitted the url (and
// firmware name) are looked up from the firmware repository for the given
// firmware + version.
package install

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/firmware/firmwareversions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
	"github.com/tidwall/sjson"
)

// InstallCmd command
type InstallCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewInstallCmd creates a command to Install firmware version on a device
func NewInstallCmd(f *cmdutil.Factory) *InstallCmd {
	ccmd := &InstallCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install firmware version on a device",
		Long:  `Install firmware version on a device`,
		Example: heredoc.Doc(`
$ c8y firmware versions install --device 1234 --firmware linux-iot --version 1.0.0
Install a firmware version (lookup url automatically).
If the firmware/version exists in the firmware repository, then it will add the url automatically


$ c8y firmware versions install --device 1234 --firmware linux-iot --version 1.0.0 --url "https://my.blobstore.com/linux-iot.tar.gz"
Install a firmware version with an explicit url
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("device", "", "Device or agent where the firmware should be installed (accepts pipeline)")
	cmd.Flags().String("firmware", "", "Firmware name (required)")
	cmd.Flags().String("version", "", "Firmware version")
	cmd.Flags().String("url", "", "Firmware url. Leave blank to automatically set it if a matching firmware/version is found in the c8y firmware repository")
	cmd.Flags().String("description", "", "Operation description")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithFirmware("firmware", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithFirmwareVersion("version", "firmware", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("device", "deviceId", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Install-FirmwareVersion"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.operation+json", ""),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("firmware")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *InstallCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.Input(); err != nil {
		return err
	}

	err = r.Body(
		flags.WithOverrideValue("device", "deviceId"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("device", "deviceId"),
		flags.WithStringValue("firmware", "c8y_Firmware.name"),
		flags.WithStringValue("version", "c8y_Firmware.version"),
		flags.WithStringValue("url", "c8y_Firmware.url"),
		flags.WithStringValue("description", "description"),
		flags.WithDefaultTemplateString(`
{
  _version:: if std.objectHas(self.c8y_Firmware, 'version') then self.c8y_Firmware.version else '',
  description:
    ('Update firmware to: "%s"' % self.c8y_Firmware.name)
    + (if self._version != "" then " (%s)" % self._version else "")
}
`),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("deviceId", "c8y_Firmware.name", "c8y_Firmware.version"),
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
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		if body, err = in.ResolveBodyRef(body, "deviceId", resolveDevice); err != nil {
			return nil, err
		}

		// Look up the binary url (and firmware name) from the repository when no
		// explicit url is provided, matching v1.
		if in.String("url") == "" {
			ref := versionRef(in.String("version"), in.String("firmware"))
			vres := client.Repository.Firmware.Versions.Get(in.ResolveContext(), ref, firmwareversions.GetOptions{WithParents: true})
			if vres.Err != nil {
				return nil, vres.Err
			}
			if url := vres.Data.URL(); url != "" {
				if body, err = sjson.SetBytes(body, "c8y_Firmware.url", url); err != nil {
					return nil, err
				}
			}
			if name := vres.Data.FirmwareName(); name != "" {
				if body, err = sjson.SetBytes(body, "c8y_Firmware.name", name); err != nil {
					return nil, err
				}
			}
		}

		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Operation] {
				return client.Operations.CreateRaw(ctx, body)
			})
		}, nil
	})
}

// versionRef builds the version resolver reference for the firmware repository
// lookup: a plain managed-object id (the version itself) is used directly,
// otherwise the version is scoped to its firmware (by id or name).
func versionRef(version, firmware string) string {
	if version == "" || firmware == "" || c8ydata.IsID(version) {
		return version
	}
	if c8ydata.IsID(firmware) {
		return firmwareversions.NewRef().ByVersion(version, firmware)
	}
	return firmwareversions.NewRef().ByVersionAndName(version, firmware)
}
