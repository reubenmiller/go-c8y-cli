package devicegroups

import (
	cmdAssignDevice "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/assigndevice"
	cmdAssignGroup "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/assigngroup"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/list"
	cmdListAssets "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/listassets"
	cmdUnassignDevice "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/unassigndevice"
	cmdUnassignGroup "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/unassigngroup"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdDevicegroups struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdDevicegroups {
	ccmd := &SubCmdDevicegroups{}

	cmd := &cobra.Command{
		Use:   "devicegroups",
		Short: "Cumulocity device groups",
		Long:  `REST endpoint to interact with Cumulocity device groups`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdAssignDevice.NewAssignDeviceCmd(f).GetCommand())
	cmd.AddCommand(cmdAssignGroup.NewAssignGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassignDevice.NewUnassignDeviceCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassignGroup.NewUnassignGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdListAssets.NewListAssetsCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
