package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type MeasurementsCmd struct {
	*subcommand.SubCommand
}

func NewMeasurementsRootCmd() *MeasurementsCmd {
	ccmd := &MeasurementsCmd{}

	cmd := &cobra.Command{
		Use:   "measurements",
		Short: "Cumulocity measurements",
		Long:  `REST endpoint to interact with Cumulocity measurements`,
	}

	// Subcommands
	cmd.AddCommand(NewGetMeasurementCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetMeasurementSeriesCmd().GetCommand())
	cmd.AddCommand(NewGetMeasurementCmd().GetCommand())
	cmd.AddCommand(NewNewMeasurementCmd().GetCommand())
	cmd.AddCommand(NewDeleteMeasurementCmd().GetCommand())
	cmd.AddCommand(NewDeleteMeasurementCollectionCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
