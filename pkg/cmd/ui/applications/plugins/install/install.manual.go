package install

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

func NewCmd(f *cmdutil.Factory) *plugins.PluginCmd {
	ccmd := &plugins.PluginCmd{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install UI plugins in an application",
		Long: heredoc.Doc(`
			Install UI plugins in an application
		`),
		Example: heredoc.Doc(`
			$ c8y ui applications plugins install --application devicemanagement --plugin myplugin@latest --plugin someother@1.2.3
			Install multiple UI plugins to the devicemanagement application

			$ c8y ui applications plugins install --application devicemanagement --plugin myplugin --template "{config:{remotes:{'other@1.0.0':[]}}}"
			Install myplugin via a lookup and add manual configuration using templates (for power users only!)
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return plugins.NewPluginRunner(cmd, args, f, ccmd)()
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.Application, "application", "", "Application (required) (accepts pipeline)")
	cmd.Flags().StringSliceVar(&ccmd.Add, "plugin", []string{}, "UI plugin to be installed")

	completion.WithOptions(
		cmd,
		completion.WithHostedApplication("application", func() (*c8y.Client, error) { return ccmd.Factory.Client() }),
		completion.WithUIPluginWithVersions("plugin", func() (*c8y.Client, error) { return ccmd.Factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("application", "application", false, "id"),
		flags.WithSemanticMethod("POST"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
