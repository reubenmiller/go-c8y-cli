package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type IdentityCmd struct {
	*subcommand.SubCommand
}

func NewIdentityRootCmd() *IdentityCmd {
	ccmd := &IdentityCmd{}

	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Cumulocity identity",
		Long:  `REST endpoint to interact with Cumulocity identity objects`,
	}

	// Subcommands
	cmd.AddCommand(NewGetExternalIDCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetExternalIDCmd().GetCommand())
	cmd.AddCommand(NewDeleteExternalIDCmd().GetCommand())
	cmd.AddCommand(NewNewExternalIDCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
