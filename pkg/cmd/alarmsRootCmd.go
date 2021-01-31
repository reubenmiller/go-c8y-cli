package cmd

import (
	"github.com/spf13/cobra"
)

type AlarmsCmd struct {
	*baseCmd
}

func NewAlarmsRootCmd() *AlarmsCmd {
	ccmd := &AlarmsCmd{}

	cmd := &cobra.Command{
		Use:   "alarms",
		Short: "Cumulocity alarms",
		Long:  `REST endpoint to interact with Cumulocity alarms`,
	}

	// Subcommands
	cmd.AddCommand(NewGetAlarmCollectionCmd().getCommand())
	cmd.AddCommand(NewNewAlarmCmd().getCommand())
	cmd.AddCommand(NewUpdateAlarmCollectionCmd().getCommand())
	cmd.AddCommand(NewGetAlarmCmd().getCommand())
	cmd.AddCommand(NewUpdateAlarmCmd().getCommand())
	cmd.AddCommand(NewDeleteAlarmCollectionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
