package cmd

import (
	"github.com/spf13/cobra"
)

type eventsCmd struct {
	*baseCmd
}

func newEventsRootCmd() *eventsCmd {
	ccmd := &eventsCmd{}

	cmd := &cobra.Command{
		Use:   "events",
		Short: "Cumulocity events",
		Long:  `REST endpoint to interact with Cumulocity events`,
	}

	// Subcommands
	cmd.AddCommand(newGetEventCollectionCmd().getCommand())
	cmd.AddCommand(newDeleteEventCollectionCmd().getCommand())
	cmd.AddCommand(newGetEventCmd().getCommand())
	cmd.AddCommand(newNewEventCmd().getCommand())
	cmd.AddCommand(newUpdateEventCmd().getCommand())
	cmd.AddCommand(newDeleteEventCmd().getCommand())
	cmd.AddCommand(newGetEventBinaryCmd().getCommand())
	cmd.AddCommand(newNewEventBinaryCmd().getCommand())
	cmd.AddCommand(newUpdateEventBinaryCmd().getCommand())
	cmd.AddCommand(newDeleteEventBinaryCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
