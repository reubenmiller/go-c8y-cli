package cmd

import (
	"github.com/spf13/cobra"
)

type EventsCmd struct {
	*baseCmd
}

func NewEventsRootCmd() *EventsCmd {
	ccmd := &EventsCmd{}

	cmd := &cobra.Command{
		Use:   "events",
		Short: "Cumulocity events",
		Long:  `REST endpoint to interact with Cumulocity events`,
	}

	// Subcommands
	cmd.AddCommand(NewGetEventCollectionCmd().getCommand())
	cmd.AddCommand(NewDeleteEventCollectionCmd().getCommand())
	cmd.AddCommand(NewGetEventCmd().getCommand())
	cmd.AddCommand(NewNewEventCmd().getCommand())
	cmd.AddCommand(NewUpdateEventCmd().getCommand())
	cmd.AddCommand(NewDeleteEventCmd().getCommand())
	cmd.AddCommand(NewGetEventBinaryCmd().getCommand())
	cmd.AddCommand(NewNewEventBinaryCmd().getCommand())
	cmd.AddCommand(NewUpdateEventBinaryCmd().getCommand())
	cmd.AddCommand(NewDeleteEventBinaryCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
