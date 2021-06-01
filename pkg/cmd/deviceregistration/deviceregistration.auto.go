package deviceregistration

import (
	cmdApprove "github.com/reubenmiller/go-c8y-cli/pkg/cmd/deviceregistration/approve"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/deviceregistration/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/deviceregistration/get"
	cmdGetCredentials "github.com/reubenmiller/go-c8y-cli/pkg/cmd/deviceregistration/getcredentials"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/deviceregistration/list"
	cmdRegister "github.com/reubenmiller/go-c8y-cli/pkg/cmd/deviceregistration/register"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdDeviceregistration struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdDeviceregistration {
	ccmd := &SubCmdDeviceregistration{}

	cmd := &cobra.Command{
		Use:   "deviceregistration",
		Short: "Cumulocity device credentials",
		Long:  `REST endpoint to interact with Cumulocity device credentials api`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdRegister.NewRegisterCmd(f).GetCommand())
	cmd.AddCommand(cmdApprove.NewApproveCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdGetCredentials.NewGetCredentialsCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
