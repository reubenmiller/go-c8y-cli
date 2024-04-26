package list

import (
	"context"
	"encoding/json"
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

// Custom plugin output format
type PluginReference struct {
	ID          string         `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Modules     []string       `json:"modules,omitempty"`
	ContextPath string         `json:"contextPath,omitempty"`
	Description string         `json:"description,omitempty"`
	Version     string         `json:"version,omitempty"`
	Plugin      map[string]any `json:"plugin,omitempty"`
}

func NewPluginReference(app gjson.Result, version string, modules []string) (*PluginReference, error) {

	plugin := make(map[string]any)
	jsonErr := json.Unmarshal([]byte(app.Raw), &plugin)

	return &PluginReference{
		ID:          app.Get("id").String(),
		Name:        app.Get("name").String(),
		Modules:     modules,
		ContextPath: app.Get("contextPath").String(),
		Description: app.Get("manifest.description").String(),
		Version:     version,
		Plugin:      plugin,
	}, jsonErr
}

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

	cmd.Flags().StringVar(&ccmd.Application, "application", "", "Application (required) (accepts pipeline)")

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
	pluginModules := make(map[string][]string)
	if v, ok := matches[0].Data.Value.(gjson.Result); ok {
		if remotes := v.Get("config.remotes"); remotes.Exists() {
			remotes.ForEach(func(key, value gjson.Result) bool {
				pluginKey := key.String()
				pluginRemotes = append(pluginRemotes, pluginKey)

				if value.IsArray() {
					pluginModules[pluginKey] = make([]string, 0)
					value.ForEach(func(i, module gjson.Result) bool {
						moduleName := module.String()
						if moduleName != "" {
							pluginModules[pluginKey] = append(pluginModules[pluginKey], moduleName)
						}
						return true
					})
				}
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
	// Output plugins (if found) as a custom data model (which reflects a similar view as presented in the UI)
	//
	for _, contextVersion := range pluginRemotes {
		contextPath, version, ok := strings.Cut(contextVersion, "@")
		if ok {
			var reference *PluginReference
			var referenceErr error
			if i, found := localPluginLookup[contextPath]; found {
				reference, referenceErr = NewPluginReference(localPlugins.Items[i], version, pluginModules[contextVersion])
			} else if i, found := sharedPluginLookup[contextPath]; found {
				reference, referenceErr = NewPluginReference(sharedPlugins.Items[i], version, pluginModules[contextVersion])
			} else {
				if !dryRun {
					log.Warnf("could not find plugin. contextPath=%s, version=%s", contextPath, version)
				}
			}

			if referenceErr != nil {
				return referenceErr
			}

			if reference != nil {
				referenceJSON, err := json.Marshal(reference)
				if err != nil {
					return err
				}

				if err := n.factory.WriteJSONToConsole(cfg, cmd, "", referenceJSON); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
