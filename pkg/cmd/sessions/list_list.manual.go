package sessions

import (
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/sessions/login"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/sessions/selectsession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/spf13/cobra"
)

type CmdList struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	sessionFilter string
	fromExtension string
}

func NewCmdList(f *cmdutil.Factory) *CmdList {
	ccmd := &CmdList{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get a Cumulocity session",
		Long:  `Get a Cumulocity session`,
		Example: heredoc.Doc(`
			Example 1: Show an interactive list of all available sessions

			#> c8y sessions list

			Example 2: Select a session and filter the selection of session by the name "customer"

			#> export C8Y_SESSION=$( c8y sessions list --sessionFilter "customer" )

			Example 3: Print the sessions of the extension session provider "bitwarden" as json (without secrets). Arguments after "--" are passed to the provider

			#> c8y sessions list --from-extension bitwarden -- --folder c8y
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.sessionFilter, "sessionFilter", "", "Filter to be applied to the list of sessions even before the values can be selected")

	cmd.Flags().StringVar(&ccmd.fromExtension, "from-extension", "", "Print the sessions of an extension session provider as json, i.e. <provider> or <extension>/<provider>. The provider must declare the list and select commands")
	_ = cmd.RegisterFlagCompletionFunc("from-extension", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		providers := []extensions.InstalledSessionProvider{}
		for _, p := range extensions.ListSessionProviders(f.ExtensionManager().List(), nil) {
			if p.Provider.CanList() {
				providers = append(providers, p)
			}
		}
		return extensions.SessionProviderCompletions(providers), cobra.ShellCompDirectiveNoFileComp
	})

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdList) RunE(cmd *cobra.Command, args []string) error {
	if n.fromExtension != "" {
		cfg, err := n.factory.Config()
		if err != nil {
			return err
		}
		_, sessions, err := login.ListExtensionSessions(cmd.Context(), n.factory, n.fromExtension, args, cfg.SessionProviderSecrets())
		if err != nil {
			return err
		}
		out, err := json.MarshalIndent(sessions, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintf(n.factory.IOStreams.Out, "%s\n", out)
		return nil
	}

	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}

	sessionFile, err := selectsession.SelectSession(n.factory.IOStreams, cfg, log, n.sessionFilter)

	if err != nil {
		return err
	}
	fmt.Printf("%s\n", sessionFile)
	return nil
}
