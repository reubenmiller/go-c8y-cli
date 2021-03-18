package cmd

import (
	cmdApproveDeviceRequest "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicecredentials/approvedevicerequest"
	cmdDeleteNewDeviceRequest "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicecredentials/deletenewdevicerequest"
	cmdGetNewDeviceRequest "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicecredentials/getnewdevicerequest"
	cmdListNewDeviceRequests "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicecredentials/listnewdevicerequests"
	cmdRegisterNewDevice "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicecredentials/registernewdevice"
	cmdRequestDeviceCredentials "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicecredentials/requestdevicecredentials"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdDevicecredentials struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdDevicecredentials {
	ccmd := &SubCmdDevicecredentials{}

	cmd := &cobra.Command{
		Use:   "devicecredentials",
		Short: "Cumulocity device credentials",
		Long:  `REST endpoint to interact with Cumulocity device credentials api`,
	}

	// Subcommands
	cmd.AddCommand(cmdListNewDeviceRequests.NewListNewDeviceRequestsCmd(f).GetCommand())
	cmd.AddCommand(cmdGetNewDeviceRequest.NewGetNewDeviceRequestCmd(f).GetCommand())
	cmd.AddCommand(cmdRegisterNewDevice.NewRegisterNewDeviceCmd(f).GetCommand())
	cmd.AddCommand(cmdApproveDeviceRequest.NewApproveDeviceRequestCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteNewDeviceRequest.NewDeleteNewDeviceRequestCmd(f).GetCommand())
	cmd.AddCommand(cmdRequestDeviceCredentials.NewRequestDeviceCredentialsCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
