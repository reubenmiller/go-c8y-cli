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

func DeleteNewCmd(f *cmdutil.Factory) *plugins.PluginCmd {
	ccmd := &plugins.PluginCmd{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Remove UI plugins from an application",
		Long: heredoc.Doc(`
			Remove UI plugins assigned to an application
		`),
		Example: heredoc.Doc(`
			$ c8y ui applications plugins delete --application devicemanagement --plugin myext@latest --plugin someother@1.2.3
			Delete multiple UI plugins from an application

			$ c8y ui applications plugins delete --application devicemanagement --invalid
			Delete orphaned or revoked UI plugins from an application

			$ c8y ui applications plugins delete --application devicemanagement --all
			Delete all UI plugins from an application
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return ccmd.Run(cmd)
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.Application, "application", "", "Application (required) (accepts pipeline)")
	cmd.Flags().StringSliceVar(&ccmd.Remove, "plugin", []string{}, "UI plugin to be removed")
	cmd.Flags().BoolVar(&ccmd.ReplaceAll, "all", false, "Delete all UI plugins")
	cmd.Flags().BoolVar(&ccmd.RemoveInvalid, "invalid", false, "Remove orphaned or revoked plugins")

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
		flags.WithSemanticMethod("DELETE"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
