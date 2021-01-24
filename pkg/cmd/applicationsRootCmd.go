package cmd

import (
	"github.com/spf13/cobra"
)

type applicationsCmd struct {
	*baseCmd
}

func newApplicationsRootCmd() *applicationsCmd {
	ccmd := &applicationsCmd{}

	cmd := &cobra.Command{
		Use:   "applications",
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
	cmd.AddCommand(newGetApplicationBinaryCollectionCmd().getCommand())
	cmd.AddCommand(newDeleteApplicationBinaryCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
