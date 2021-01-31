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
	cmd.AddCommand(NewGetMeasurementCmd().getCommand())
	cmd.AddCommand(NewGetMeasurementCollectionCmd().getCommand())
	cmd.AddCommand(NewGetMeasurementSeriesCmd().getCommand())
	cmd.AddCommand(NewDeleteMeasurementCmd().getCommand())
	cmd.AddCommand(NewDeleteMeasurementCollectionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
