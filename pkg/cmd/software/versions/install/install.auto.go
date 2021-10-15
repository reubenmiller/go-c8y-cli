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

// NewInstallCmd creates a command to Install software version on a device
func NewInstallCmd(f *cmdutil.Factory) *InstallCmd {
	ccmd := &InstallCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install software version on a device",
		Long:  `Install software version on a device`,
		Example: heredoc.Doc(`
$ c8y software versions install --device 1234 --software go-c8y-cli --version 1.0.0
Install a software package version
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device or agent where the software should be installed (accepts pipeline)")
	cmd.Flags().String("software", "", "Software name (required)")
	cmd.Flags().String("version", "", "Software version id or name")
	cmd.Flags().String("url", "", "Software url")
	cmd.Flags().String("description", "Install software package", "Operation description")
	cmd.Flags().String("action", "install", "Software action")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithSoftware("software", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithSoftwareVersion("version", "software", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("action", "install"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),

		flags.WithExtendedPipelineSupport("device", "deviceId", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("software")

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
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithDataFlagValue(),
		c8yfetcher.WithDeviceByNameFirstMatch(client, args, "device", "deviceId"),
		flags.WithStringValue("software", "c8y_SoftwareUpdate.0.name"),
		flags.WithStringValue("version", "c8y_SoftwareUpdate.0.version"),
		flags.WithStringValue("url", "c8y_SoftwareUpdate.0.url"),
		flags.WithStringValue("description", "description"),
		c8yfetcher.WithSoftwareVersionData(client, "software", "version", "url", args, "", "c8y_SoftwareUpdate.0"),
		flags.WithStringValue("action", "c8y_SoftwareUpdate.0.action"),
		cmdutil.WithTemplateValue(cfg),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("deviceId", "c8y_SoftwareUpdate.0.name", "c8y_SoftwareUpdate.0.version", "c8y_SoftwareUpdate.0.action"),
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
