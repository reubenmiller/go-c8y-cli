package open

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
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
		completion.WithApplicationContext("application", ccmd.factory.ClientV2),
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
		completion.WithDevice("device", ccmd.factory.ClientV2),
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
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	if err := r.InputFlag("device"); err != nil {
		return err
	}
	client, err := r.Client()
	if err != nil {
		return err
	}
	cfg := r.Config
	baseURL := strings.TrimSuffix(client.BaseURL.String(), "/")

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		device := in.String("device")
		return func(ctx context.Context) any {
			if device != "" {
				id, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(device), nil)
				if err != nil {
					return err
				}
				device = id
			}
			target, err := url.Parse(baseURL + "/" + strings.TrimPrefix(n.pagePath(device), "/"))
			if err != nil {
				return err
			}
			switch {
			case n.noBrowser:
				return c8ystream.FromDocs(jsondoc.New([]byte(target.String())))
			case cfg.DryRun():
				fmt.Fprintf(n.factory.IOStreams.Out, "WHATIF: open %s in default browser\n", target)
				return nil
			}
			cfg.Logger.Infof("Opening web app: %s", target)
			return n.factory.Browser.Browse(target.String())
		}, nil
	})
}

// pagePath returns the application path for the device (which may be empty),
// from --path or a default for the application.
func (n *OpenCmd) pagePath(device string) string {
	template := n.path
	if template == "" {
		switch {
		case strings.Contains(n.application, "devicemanagement"):
			if device != "" {
				template = "/apps/{application}/index.html#/device/{device}/{page}"
			} else {
				template = "/apps/{application}/index.html#/{page}"
			}
		case strings.Contains(n.application, "cockpit"), strings.Contains(n.application, "administration"):
			template = "/apps/{application}/index.html"
		}
	}
	return strings.NewReplacer(
		"{application}", n.application,
		"{device}", device,
		"{page}", n.page,
	).Replace(template)
}
