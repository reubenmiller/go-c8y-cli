package devices

import (
	cmdAssignChild "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/assignchild"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/get"
	cmdGetChild "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/getchild"
	cmdGetSupportedMeasurements "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/getsupportedmeasurements"
	cmdGetSupportedSeries "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/getsupportedseries"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/list"
	cmdListAssets "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/listassets"
	cmdListChildren "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/listchildren"
	cmdUnassignChild "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/unassignchild"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
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
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGetSupportedMeasurements.NewGetSupportedMeasurementsCmd(f).GetCommand())
	cmd.AddCommand(cmdGetSupportedSeries.NewGetSupportedSeriesCmd(f).GetCommand())
	cmd.AddCommand(cmdAssignChild.NewAssignChildCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassignChild.NewUnassignChildCmd(f).GetCommand())
	cmd.AddCommand(cmdListChildren.NewListChildrenCmd(f).GetCommand())
	cmd.AddCommand(cmdGetChild.NewGetChildCmd(f).GetCommand())
	cmd.AddCommand(cmdListAssets.NewListAssetsCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
