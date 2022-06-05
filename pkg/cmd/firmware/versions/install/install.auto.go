// Code generated from specification version 1.0.0: DO NOT EDIT
package install

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
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
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device or agent where the firmware should be installed (accepts pipeline)")
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
	)

	// Required flags
	_ = cmd.MarkFlagRequired("firmware")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *InstallCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"),
		flags.WithProcessingModeValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder(true)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithDataFlagValue(),
		c8yfetcher.WithDeviceByNameFirstMatch(client, args, "device", "deviceId"),
		flags.WithStringValue("firmware", "c8y_Firmware.name"),
		flags.WithStringValue("version", "c8y_Firmware.version"),
		flags.WithStringValue("url", "c8y_Firmware.url"),
		c8yfetcher.WithFirmwareVersionData(client, "firmware", "version", "url", args, "", "c8y_Firmware"),
		flags.WithStringValue("description", "description"),
		flags.WithDefaultTemplateString(`
{
  _version:: if std.objectHas(self.c8y_Firmware, 'version') then self.c8y_Firmware.version else '',
  description:
    ('Update firmware to: "%s"' % self.c8y_Firmware.name)
    + (if self._version != "" then " (%s)" % self._version else "")
}
`),
		cmdutil.WithTemplateValue(cfg),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("deviceId", "c8y_Firmware.name", "c8y_Firmware.version"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("devicecontrol/operations")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "POST",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
