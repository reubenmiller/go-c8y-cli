package cmd

import (
	"github.com/spf13/cobra"
)

type ApplicationsCmd struct {
	*baseCmd
}

func NewApplicationsRootCmd() *ApplicationsCmd {
	ccmd := &ApplicationsCmd{}

	cmd := &cobra.Command{
		Use:   "applications",
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
	cmd.AddCommand(NewGetApplicationBinaryCollectionCmd().getCommand())
	cmd.AddCommand(NewDeleteApplicationBinaryCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
