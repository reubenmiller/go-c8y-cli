package cmd

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/measurements/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/measurements/delete"
	cmdDeleteCollection "github.com/reubenmiller/go-c8y-cli/pkg/cmd/measurements/deletecollection"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/measurements/get"
	cmdGetSeries "github.com/reubenmiller/go-c8y-cli/pkg/cmd/measurements/getseries"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/measurements/list"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdMeasurements struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdMeasurements {
	ccmd := &SubCmdMeasurements{}

	cmd := &cobra.Command{
		Use:   "measurements",
		Short: "Cumulocity measurements",
		Long:  `REST endpoint to interact with Cumulocity measurements`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGetSeries.NewGetSeriesCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteCollection.NewDeleteCollectionCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
