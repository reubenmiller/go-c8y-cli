package cmd

import (
	"github.com/spf13/cobra"
)

type measurementCmd struct {
	*baseCmd
}

func newMeasurementRootCmd() *measurementCmd {
	ccmd := &measurementCmd{}

	cmd := &cobra.Command{
		Use:   "measurement",
		Short: "Measurement REST endpoint",
		Long: `

		`,
	}

	// Subcommands
	cmd.AddCommand(newGetMeasurementCmd().getCommand())
	cmd.AddCommand(newGetMeasurementCollectionCmd().getCommand())
	cmd.AddCommand(newGetMeasurementSeriesCmd().getCommand())
	cmd.AddCommand(newDeleteMeasurementCmd().getCommand())
	cmd.AddCommand(newDeleteMeasurementCollectionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
