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
	cmd.AddCommand(newGetApplicationCollectionCmd().getCommand())
	cmd.AddCommand(newNewApplicationCmd().getCommand())
	cmd.AddCommand(newCopyApplicationCmd().getCommand())
	cmd.AddCommand(newGetApplicationCmd().getCommand())
	cmd.AddCommand(newDeleteApplicationCmd().getCommand())
	cmd.AddCommand(newUpdateApplicationCmd().getCommand())
	cmd.AddCommand(newNewApplicationBinaryCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
