package cmd

import (
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
	cmd.AddCommand(newNewSessionCmd().getCommand())
	cmd.AddCommand(newGetSessionCmd().getCommand())
	cmd.AddCommand(newEncryptTextCmd().getCommand())
	cmd.AddCommand(newDecryptTextCmd().getCommand())
	cmd.AddCommand(newCheckSessionPassphraseCmd().getCommand())
	cmd.AddCommand(newListSessionCmd().getCommand())
	cmd.AddCommand(newSessionLoginCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
