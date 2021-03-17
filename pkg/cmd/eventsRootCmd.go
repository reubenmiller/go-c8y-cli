package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type EventsCmd struct {
	*subcommand.SubCommand
}

func NewEventsRootCmd() *EventsCmd {
	ccmd := &EventsCmd{}

	cmd := &cobra.Command{
		Use:   "events",
		Short: "Cumulocity events",
		Long:  `REST endpoint to interact with Cumulocity events`,
	}

	// Subcommands
	cmd.AddCommand(NewGetEventCollectionCmd().GetCommand())
	cmd.AddCommand(NewDeleteEventCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetEventCmd().GetCommand())
	cmd.AddCommand(NewNewEventCmd().GetCommand())
	cmd.AddCommand(NewUpdateEventCmd().GetCommand())
	cmd.AddCommand(NewDeleteEventCmd().GetCommand())
	cmd.AddCommand(NewGetEventBinaryCmd().GetCommand())
	cmd.AddCommand(NewNewEventBinaryCmd().GetCommand())
	cmd.AddCommand(NewUpdateEventBinaryCmd().GetCommand())
	cmd.AddCommand(NewDeleteEventBinaryCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
