package plugins

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yextensions/plugins"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/spf13/cobra"
)

func ReplaceNewCmd(f *cmdutil.Factory) *plugins.PluginCmd {
	ccmd := &plugins.PluginCmd{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "replace",
		Short: "Replace UI plugins in an application",
		Long: heredoc.Doc(`
			Replace all existing ui plugins with a new set of plugins in an application
		`),
		Example: heredoc.Doc(`
			$ c8y ui applications plugins replace --application devicemanagement --plugin myext@latest --plugin someother@1.2.3
			Replace all existing plugins with a new list of plugins
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return ccmd.Run(cmd)
	}

	cmd.SilenceUsage = true

	ccmd.ReplaceAll = true
	cmd.Flags().StringVar(&ccmd.Application, "application", "", "Application (required) (accepts pipeline)")
	cmd.Flags().StringSliceVar(&ccmd.Add, "plugin", []string{}, "UI plugin to be installed")

	completion.WithOptions(
		cmd,
		completion.WithHostedApplication("application", ccmd.Factory.ClientV2),
		completion.WithUIPluginWithVersions("plugin", ccmd.Factory.ClientV2),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("application", "application", false, "id"),
		flags.WithSemanticMethod("PUT"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
