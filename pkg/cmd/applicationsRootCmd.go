package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type ApplicationsCmd struct {
	*subcommand.SubCommand
}

func NewApplicationsRootCmd() *ApplicationsCmd {
	ccmd := &ApplicationsCmd{}

	cmd := &cobra.Command{
		Use:   "applications",
		Short: "Cumulocity applications",
		Long:  `REST endpoint to interact with Cumulocity applications`,
	}

	// Subcommands
	cmd.AddCommand(NewGetApplicationCollectionCmd().GetCommand())
	cmd.AddCommand(NewNewApplicationCmd().GetCommand())
	cmd.AddCommand(NewCopyApplicationCmd().GetCommand())
	cmd.AddCommand(NewGetApplicationCmd().GetCommand())
	cmd.AddCommand(NewDeleteApplicationCmd().GetCommand())
	cmd.AddCommand(NewUpdateApplicationCmd().GetCommand())
	cmd.AddCommand(NewNewApplicationBinaryCmd().GetCommand())
	cmd.AddCommand(NewGetApplicationBinaryCollectionCmd().GetCommand())
	cmd.AddCommand(NewDeleteApplicationBinaryCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
