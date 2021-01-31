package cmd

import (
	"github.com/spf13/cobra"
)

type idsCmd struct {
	*baseCmd
}

func newIdsRootCmd() *idsCmd {
	ccmd := &idsCmd{}

	cmd := &cobra.Command{
		Use:   "ids",
		Short: "Cumulocity applications",
		Long:  `REST endpoint to interact with Cumulocity applications`,
	}

	// Subcommands
	cmd.AddCommand(NewGetApplicationCollectionCmd().getCommand())
	cmd.AddCommand(NewNewApplicationCmd().getCommand())
	cmd.AddCommand(NewCopyApplicationCmd().getCommand())
	cmd.AddCommand(NewGetApplicationCmd().getCommand())
	cmd.AddCommand(NewDeleteApplicationCmd().getCommand())
	cmd.AddCommand(NewUpdateApplicationCmd().getCommand())
	cmd.AddCommand(NewNewApplicationBinaryCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
