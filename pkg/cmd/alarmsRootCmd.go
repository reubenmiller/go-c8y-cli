package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type AlarmsCmd struct {
	*subcommand.SubCommand
}

func NewAlarmsRootCmd() *AlarmsCmd {
	ccmd := &AlarmsCmd{}

	cmd := &cobra.Command{
		Use:   "alarms",
		Short: "Cumulocity alarms",
		Long:  `REST endpoint to interact with Cumulocity alarms`,
	}

	// Subcommands
	cmd.AddCommand(NewGetAlarmCollectionCmd().GetCommand())
	cmd.AddCommand(NewNewAlarmCmd().GetCommand())
	cmd.AddCommand(NewUpdateAlarmCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetAlarmCmd().GetCommand())
	cmd.AddCommand(NewUpdateAlarmCmd().GetCommand())
	cmd.AddCommand(NewDeleteAlarmCollectionCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
