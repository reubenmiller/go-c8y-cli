package cmd

import (
	"github.com/spf13/cobra"
)

type measurementsCmd struct {
	*baseCmd
}

func newMeasurementsRootCmd() *measurementsCmd {
	ccmd := &measurementsCmd{}

	cmd := &cobra.Command{
		Use:   "measurements",
		Short: "Cumulocity measurements",
		Long:  `REST endpoint to interact with Cumulocity measurements`,
	}

	// Subcommands
	cmd.AddCommand(newGetMeasurementCollectionCmd().getCommand())
	cmd.AddCommand(newGetMeasurementSeriesCmd().getCommand())
	cmd.AddCommand(newGetMeasurementCmd().getCommand())
	cmd.AddCommand(newNewMeasurementCmd().getCommand())
	cmd.AddCommand(newDeleteMeasurementCmd().getCommand())
	cmd.AddCommand(newDeleteMeasurementCollectionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
