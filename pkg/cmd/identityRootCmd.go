package cmd

import (
	"github.com/spf13/cobra"
)

type IdentityCmd struct {
	*baseCmd
}

func NewIdentityRootCmd() *IdentityCmd {
	ccmd := &IdentityCmd{}

	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Cumulocity identity",
		Long:  `REST endpoint to interact with Cumulocity identity objects`,
	}

	// Subcommands
	cmd.AddCommand(NewGetExternalIDCollectionCmd().getCommand())
	cmd.AddCommand(NewGetExternalIDCmd().getCommand())
	cmd.AddCommand(NewDeleteExternalIDCmd().getCommand())
	cmd.AddCommand(NewNewExternalIDCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
