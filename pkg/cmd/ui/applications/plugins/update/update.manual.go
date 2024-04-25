package update

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yextensions/plugins"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

func NewCmdUpdate(f *cmdutil.Factory) *plugins.PluginCmd {
	ccmd := &plugins.PluginCmd{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update UI plugins in an application",
		Long: heredoc.Doc(`
			Update UI plugins in an application
		`),
		Example: heredoc.Doc(`
			$ c8y ui applications plugins update --application devicemanagement --all
			Update multiple UI plugins to the devicemanagement application

			$ c8y ui applications plugins update --application devicemanagement --plugin myext@latest
			Update specific UI plugins in an application 
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled()
		},
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return plugins.NewPluginRunner(cmd, args, f, ccmd)()
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.Application, "application", "", "Application")
	cmd.Flags().StringSliceVar(&ccmd.Add, "plugin", []string{}, "UI plugins to be update in an application")
	cmd.Flags().BoolVar(&ccmd.UpdateAll, "all", false, "Update all ui plugins to the latest version")

	completion.WithOptions(
		cmd,
		completion.WithHostedApplication("application", func() (*c8y.Client, error) { return ccmd.Factory.Client() }),
		completion.WithUIExtensionWithVersions("plugin", func() (*c8y.Client, error) { return ccmd.Factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("application", "application", false, "id"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd).SetRequiredFlags("application")

	return ccmd
}
