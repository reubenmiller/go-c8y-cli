package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type sessionsCmd struct {
	*subcommand.SubCommand
}

func NewSessionsRootCmd() *sessionsCmd {
	ccmd := &sessionsCmd{}

	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Cumulocity sessions",
		Long:  `Manage Cumulocity sessions`,
	}

	// Subcommands
	cmd.AddCommand(cmdutil.DisableAuthCheck(newNewSessionCmd().GetCommand()))
	cmd.AddCommand(newGetSessionCmd().GetCommand())
	cmd.AddCommand(cmdutil.DisableAuthCheck(NewClearSessionCmd().GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(newEncryptTextCmd().GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(newDecryptTextCmd().GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(newListSessionCmd().GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(newSessionLoginCmd().GetCommand()))

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
