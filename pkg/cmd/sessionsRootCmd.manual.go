package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type sessionsCmd struct {
	*baseCmd
}

func NewSessionsRootCmd() *sessionsCmd {
	ccmd := &sessionsCmd{}

	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Cumulocity sessions",
		Long:  `Manage Cumulocity sessions`,
	}

	// Subcommands
	cmd.AddCommand(cmdutil.DisableAuthCheck(newNewSessionCmd().getCommand()))
	cmd.AddCommand(newGetSessionCmd().getCommand())
	cmd.AddCommand(cmdutil.DisableAuthCheck(NewClearSessionCmd().getCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(newEncryptTextCmd().getCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(newDecryptTextCmd().getCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(newListSessionCmd().getCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(newSessionLoginCmd().getCommand()))

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
