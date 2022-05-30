package open

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// OpenCmd command
type OpenCmd struct {
	*subcommand.SubCommand

	application string
	page        string
	path        string
	noBrowser   bool

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

$ c8y devicegroups list | c8y applications open --path "/apps/cockpit/index.html#/group/{device}/subassets"
Open a multiple web browser tabs for a list of device groups in the cockpit application
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.application, "application", "devicemanagement", "Application name")
	cmd.Flags().StringSlice("device", []string{""}, "Device to be opened up to. Only valid if the template references {device}. (accepts pipeline)")
	cmd.Flags().StringVar(&ccmd.page, "page", "", "Device management page to open. Only valid for a specific device")
	cmd.Flags().StringVar(&ccmd.path, "path", "", "Custom path template which can reference values such as: {application}, {device}, {page}")
	cmd.Flags().BoolVar(&ccmd.noBrowser, "noBrowser", false, "Print destination URL instead of opening the browser")

	completion.WithOptions(
		cmd,
		completion.WithApplicationContext("application", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithCustomValidateSet("page", func() []string {
			cfg, err := ccmd.factory.Config()
			if err != nil {
				return []string{}
			}

			// Allow user to override the suggested pages
			pages := cfg.GetStringSlice(fmt.Sprintf("settings.application.%s.pages", ccmd.application))

			if len(pages) > 0 {
				return pages
			}

			// Use default values
			return []string{
				"alarms",
				"availability",
				"control",
				"device-configuration",
				"device-info",
				"device-profile",
				"events",
				"firmware",
				"identity",
				"logs",
				"measurements",
				"remote_access",
				"shell",
				"software",
			}
		}),
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
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

	// path parameters: Use placeholder and set template later
	path := flags.NewStringTemplate("")
	path.SetAllowEmptyValues(true)
	path.SetVariable("page", "")

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

	bounded := inputIterators.Total > 0

	// Update template based on application
	pathTemplate := ""
	switch {
	case strings.Contains(n.application, "devicemanagement"):
		if bounded {
			pathTemplate = "/apps/{application}/index.html#/device/{device}/{page}"
		} else {
			pathTemplate = "/apps/{application}/index.html#/{page}"
		}

	case strings.Contains(n.application, "cockpit"), strings.Contains(n.application, "administration"):
		pathTemplate = "/apps/{application}/index.html"
	}

	if n.path != "" {
		pathTemplate = n.path
	}

	path.SetTemplate(pathTemplate)

	for {
		currentPath, _, err := path.GetNext()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		currentURL, err := url.Parse(strings.TrimSuffix(client.BaseURL.String(), "/") + "/" + strings.TrimPrefix(string(currentPath), "/"))

		if err != nil {
			return err
		}

		if n.noBrowser {
			fmt.Fprintf(n.factory.IOStreams.Out, "%s\n", currentURL.String())
		} else if cfg.DryRun() {
			fmt.Fprintf(n.factory.IOStreams.Out, "WHATIF: open %s in default browser\n", currentURL.String())
		} else {
			cfg.Logger.Infof("Opening web app: %s", currentURL.String())
			if err := n.factory.Browser.Browse(currentURL.String()); err != nil {
				return err
			}
		}

		if !bounded {
			break
		}
	}

	return nil
}
