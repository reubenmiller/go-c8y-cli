package open

import (
	"fmt"
	"io"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// OpenCmd command
type OpenCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewOpenCmd creates a command to Get application
func NewOpenCmd(f *cmdutil.Factory) *OpenCmd {
	ccmd := &OpenCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "open",
		Short: "Open application in a web browser",
		Long:  `Open application in a web browser`,
		Example: heredoc.Doc(`
$ c8y applications open --application cockpit
Open the cockpit application in a local web browser

$ c8y devices list | c8y applications open --application devicemanagement --page control
Open a multiple web browser tabs in the devicemanagement application, one for each device found
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("application", "devicemanagement", "Application name, defaults to")
	cmd.Flags().StringSlice("device", []string{""}, "The ManagedObject which is the source of this event. (accepts pipeline)")
	cmd.Flags().String("page", "device-info", "Device management page to open. Only valid for a specific device")
	cmd.Flags().String("path", "", "Custom path template which can reference values such as: {application}, {device}, {page}")

	completion.WithOptions(
		cmd,
		completion.WithHostedApplication("application", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("page", "device-info", "measurements", "alarms", "control", "availability", "events", "identity"),
	)

	flags.WithOptions(
		cmd,

		flags.WithExtendedPipelineSupport("device", "device", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	cmdutil.DisableAuthCheck(cmd)

	// Required flags
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *OpenCmd) RunE(cmd *cobra.Command, args []string) error {
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

	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParameters(query)

	application, err := cmd.Flags().GetString("application")
	if err != nil {
		return err
	}

	pathTemplate := ""
	switch application {
	case "devicemanagement":
		pathTemplate = "/app/{application}/index.html#/device/{device}/{page}"

	case "cockpit":
		fallthrough
	case "administration":
		pathTemplate = "/app/{application}/index.html"
	}

	if cmd.Flags().Changed("page") {
		customPath, err := cmd.Flags().GetString("page")
		if err != nil {
			return err
		}
		pathTemplate = customPath
	}

	// path parameters
	path := flags.NewStringTemplate(pathTemplate)
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		flags.WithStringValue("page", "page"),
		flags.WithStringValue("application", "application"),
		c8yfetcher.WithDeviceByNameFirstMatch(client, args, "device", "device"),
	)
	if err != nil {
		return err
	}

	bounded := path.IsBound()
	for {
		url, _, err := path.GetNext()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		req, err := client.NewRequest("GET", string(url), "", nil)

		if err != nil {
			return err
		}

		if cfg.DryRun() {
			fmt.Fprintf(n.factory.IOStreams.Out, "open %s in browser", req.URL.String())
		} else {
			if err := n.factory.Browser.Browse(req.URL.String()); err != nil {
				return err
			}
		}

		if !bounded {
			break
		}
	}

	return nil
}
