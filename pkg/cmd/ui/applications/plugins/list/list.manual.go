package list

import (
	"context"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type CmdList struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	Application string
}

func NewCmd(f *cmdutil.Factory) *CmdList {
	ccmd := &CmdList{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed UI plugins",
		Long: heredoc.Doc(`
			List UI plugins which are installed in an application
		`),
		Example: heredoc.Doc(`
			$ c8y ui applications plugins list --application devicemanagement
			List the ui plugins which are installed
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.Application, "application", "", "Application")

	completion.WithOptions(
		cmd,
		completion.WithHostedApplication("application", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("application", "application", false, "id"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdList) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}

	matches, err := c8yfetcher.FindHostedApplications(n.factory, []string{n.Application}, true, "", false)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return nil
	}

	pluginRemotes := make([]string, 0)
	if v, ok := matches[0].Data.Value.(gjson.Result); ok {
		if remotes := v.Get("config.remotes"); remotes.Exists() {
			remotes.ForEach(func(key, value gjson.Result) bool {
				pluginRemotes = append(pluginRemotes, key.String())
				return true
			})
		}
	}

	//
	// Build cache of plugins (to speed up lookup later)
	localPluginLookup := make(map[string]int)
	sharedPluginLookup := make(map[string]int)

	localPlugins, _, err := client.UIExtension.GetExtensions(context.Background(), &c8y.ExtensionOptions{
		Tenant:            client.TenantName,
		PaginationOptions: *c8y.NewPaginationOptions(2000),
	})
	if err != nil {
		return err
	}
	for i, plugin := range localPlugins.Applications {
		localPluginLookup[plugin.ContextPath] = i
	}

	sharedPlugins, _, err := client.UIExtension.GetExtensions(context.Background(), &c8y.ExtensionOptions{
		PaginationOptions: *c8y.NewPaginationOptions(2000),
		Availability:      c8y.ApplicationAvailabilityShared,
	})
	if err != nil {
		return err
	}
	for i, plugin := range sharedPlugins.Applications {
		sharedPluginLookup[plugin.ContextPath] = i
	}

	dryRun := cfg.ShouldUseDryRun(cmd.CommandPath())

	//
	// Output plugins (if found)
	for _, contextVersion := range pluginRemotes {
		contextPath, version, ok := strings.Cut(contextVersion, "@")
		if ok {
			if i, found := localPluginLookup[contextPath]; found {
				if err := n.factory.WriteJSONToConsole(cfg, cmd, "", []byte(localPlugins.Items[i].Raw)); err != nil {
					return err
				}
			} else if i, found := sharedPluginLookup[contextPath]; found {
				if err := n.factory.WriteJSONToConsole(cfg, cmd, "", []byte(sharedPlugins.Items[i].Raw)); err != nil {
					return err
				}
			} else {
				if !dryRun {
					log.Warnf("could not find plugin. contextPath=%s, version=%s", contextPath, version)
				}
			}
		}
	}
	return nil
}
