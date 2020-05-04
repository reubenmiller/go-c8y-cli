package cmd

import (
	"github.com/spf13/cobra"
)

type identityCmd struct {
	*baseCmd
}

func newIdentityRootCmd() *identityCmd {
	ccmd := &identityCmd{}

	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Cumulocity identity",
		Long:  `REST endpoint to interact with Cumulocity identity objects`,
	}

	// Subcommands
	cmd.AddCommand(newGetExternalIDCollectionCmd().getCommand())
	cmd.AddCommand(newGetExternalIDCmd().getCommand())
	cmd.AddCommand(newDeleteExternalIDCmd().getCommand())
	cmd.AddCommand(newNewExternalIDCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
