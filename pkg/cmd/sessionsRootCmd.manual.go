package cmd

import (
	"github.com/spf13/cobra"
)

type sessionsCmd struct {
	*baseCmd
}

func newSessionsRootCmd() *sessionsCmd {
	ccmd := &sessionsCmd{}

	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Cumulocity sessions",
		Long:  `Manage Cumulocity sessions`,
	}

	// Subcommands
	cmd.AddCommand(newNewSessionCmd().getCommand())
	// cmd.AddCommand(newEncryptPasswordCmd().getCommand())
	// cmd.AddCommand(newDecryptPasswordCmd().getCommand())
	cmd.AddCommand(newListSessionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
