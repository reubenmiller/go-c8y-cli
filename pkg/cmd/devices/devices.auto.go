package cmd

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/create"
	cmdCreateGroup "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/creategroup"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/delete"
	cmdDeleteGroup "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/deletegroup"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/get"
	cmdGetGroup "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/getgroup"
	cmdGetSupportedMeasurements "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/getsupportedmeasurements"
	cmdGetSupportedOperations "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/getsupportedoperations"
	cmdGetSupportedSeries "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/getsupportedseries"
	cmdSetRequiredAvailability "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/setrequiredavailability"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/update"
	cmdUpdateGroup "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/updategroup"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdDevices struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdDevices {
	ccmd := &SubCmdDevices{}

	cmd := &cobra.Command{
		Use:   "devices",
		Short: "Cumulocity devices",
		Long:  `REST endpoint to interact with Cumulocity devices`,
	}

	// Subcommands
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGetSupportedMeasurements.NewGetSupportedMeasurementsCmd(f).GetCommand())
	cmd.AddCommand(cmdGetSupportedSeries.NewGetSupportedSeriesCmd(f).GetCommand())
	cmd.AddCommand(cmdGetSupportedOperations.NewGetSupportedOperationsCmd(f).GetCommand())
	cmd.AddCommand(cmdSetRequiredAvailability.NewSetRequiredAvailabilityCmd(f).GetCommand())
	cmd.AddCommand(cmdGetGroup.NewGetGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdateGroup.NewUpdateGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteGroup.NewDeleteGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdCreateGroup.NewCreateGroupCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
