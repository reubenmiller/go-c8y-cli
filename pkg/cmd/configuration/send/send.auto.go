// Code generated from specification version 1.0.0: DO NOT EDIT
package send

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
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
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Identifies the target device on which this operation should be performed. (accepts pipeline)")
	cmd.Flags().String("description", "", "Text description of the operation.")
	cmd.Flags().String("configurationType", "", "Configuration type. Leave blank to automatically set it if a matching configuration is found in the c8y configuration repository")
	cmd.Flags().String("url", "", "Url to the configuration. Leave blank to automatically set it if a matching configuration is found in the c8y configuration repository")
	cmd.Flags().StringSlice("configuration", []string{""}, "Configuration name or id")

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
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *SendCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	// Runtime flag options
	flags.WithOptions(
		cmd,
		flags.WithRuntimePipelineProperty(),
	)
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
		flags.WithStringValue("description", "description"),
		flags.WithStringValue("configurationType", "c8y_DownloadConfigFile.type"),
		flags.WithStringValue("url", "c8y_DownloadConfigFile.url"),
		c8yfetcher.WithConfigurationByNameFirstMatch(client, args, "configuration", "__tmp_configuration"),
		c8yfetcher.WithConfigurationFileData(client, "configuration", "configurationType", "url", args, "", "c8y_DownloadConfigFile"),
		flags.WithDefaultTemplateString(`
{
  description:
    ('Send configuration snapshot %s of configuration type %s to device' % [self.c8y_DownloadConfigFile.name, self.c8y_DownloadConfigFile.type]),
}
`),
		flags.WithRequiredTemplateString(`
{
  __tmp_configuration:: null,
  c8y_DownloadConfigFile+: {name:: null},
}
`),
		cmdutil.WithTemplateValue(n.factory, cfg),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("deviceId"),
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
