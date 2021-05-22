package list

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/sessions/selectsession"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type CmdList struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	sessionFilter string
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
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.sessionFilter, "sessionFilter", "", "Filter to be applied to the list of sessions even before the values can be selected")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdList) RunE(cmd *cobra.Command, args []string) error {
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
