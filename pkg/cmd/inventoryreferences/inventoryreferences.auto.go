package cmd

import (
	cmdAssignChildDevice "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/assignchilddevice"
	cmdAssignDeviceToGroup "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/assigndevicetogroup"
	cmdAssignGroupToGroup "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/assigngrouptogroup"
	cmdCreateChildAddition "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/createchildaddition"
	cmdCreateChildAsset "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/createchildasset"
	cmdGetChildAsset "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/getchildasset"
	cmdGetChildDevice "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/getchilddevice"
	cmdListChildAdditions "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/listchildadditions"
	cmdListChildAssets "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/listchildassets"
	cmdListChildDevices "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/listchilddevices"
	cmdUnassignAssetFromGroup "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/unassignassetfromgroup"
	cmdUnassignChildAddition "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/unassignchildaddition"
	cmdUnassignChildDevice "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/unassignchilddevice"
	cmdUnassignDeviceFromGroup "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences/unassigndevicefromgroup"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdInventoryreferences struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdInventoryreferences {
	ccmd := &SubCmdInventoryreferences{}

	cmd := &cobra.Command{
		Use:   "inventoryreferences",
		Short: "Cumulocity inventory references",
		Long:  `REST endpoint to interact with Cumulocity managed objects`,
	}

	// Subcommands
	cmd.AddCommand(cmdListChildDevices.NewListChildDevicesCmd(f).GetCommand())
	cmd.AddCommand(cmdListChildAssets.NewListChildAssetsCmd(f).GetCommand())
	cmd.AddCommand(cmdAssignChildDevice.NewAssignChildDeviceCmd(f).GetCommand())
	cmd.AddCommand(cmdAssignDeviceToGroup.NewAssignDeviceToGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdAssignGroupToGroup.NewAssignGroupToGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdCreateChildAsset.NewCreateChildAssetCmd(f).GetCommand())
	cmd.AddCommand(cmdGetChildDevice.NewGetChildDeviceCmd(f).GetCommand())
	cmd.AddCommand(cmdGetChildAsset.NewGetChildAssetCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassignChildDevice.NewUnassignChildDeviceCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassignAssetFromGroup.NewUnassignAssetFromGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassignDeviceFromGroup.NewUnassignDeviceFromGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassignAssetFromGroup.NewUnassignAssetFromGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdListChildAdditions.NewListChildAdditionsCmd(f).GetCommand())
	cmd.AddCommand(cmdCreateChildAddition.NewCreateChildAdditionCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassignChildAddition.NewUnassignChildAdditionCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
