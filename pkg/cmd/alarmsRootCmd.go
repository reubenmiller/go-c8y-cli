package cmd

import (
	"github.com/spf13/cobra"
)

type alarmsCmd struct {
	*baseCmd
}

func newAlarmsRootCmd() *alarmsCmd {
	ccmd := &alarmsCmd{}

	cmd := &cobra.Command{
		Use:   "alarms",
		Short: "Cumulocity alarms",
		Long:  `REST endpoint to interact with Cumulocity alarms`,
	}

	// Subcommands
	cmd.AddCommand(newGetAlarmCollectionCmd().getCommand())
	cmd.AddCommand(newNewAlarmCmd().getCommand())
	cmd.AddCommand(newUpdateAlarmCollectionCmd().getCommand())
	cmd.AddCommand(newGetAlarmCmd().getCommand())
	cmd.AddCommand(newUpdateAlarmCmd().getCommand())
	cmd.AddCommand(newDeleteAlarmCollectionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
