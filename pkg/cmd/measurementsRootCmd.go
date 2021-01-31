package cmd

import (
	"github.com/spf13/cobra"
)

type MeasurementsCmd struct {
	*baseCmd
}

func NewMeasurementsRootCmd() *MeasurementsCmd {
	ccmd := &MeasurementsCmd{}

	cmd := &cobra.Command{
		Use:   "measurements",
		Short: "Cumulocity measurements",
		Long:  `REST endpoint to interact with Cumulocity measurements`,
	}

	// Subcommands
	cmd.AddCommand(NewGetMeasurementCollectionCmd().getCommand())
	cmd.AddCommand(NewGetMeasurementSeriesCmd().getCommand())
	cmd.AddCommand(NewGetMeasurementCmd().getCommand())
	cmd.AddCommand(NewNewMeasurementCmd().getCommand())
	cmd.AddCommand(NewDeleteMeasurementCmd().getCommand())
	cmd.AddCommand(NewDeleteMeasurementCollectionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
