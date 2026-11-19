// Package deviceregistration wires the `c8y deviceregistration` command and its
// spec-generated subcommands. All of them are hand-written v2 c8ystream commands
// (backed by the go-c8y v2 Devices.Registration service), so this group command
// is itself hand-written rather than generated. The bulk register-* commands are
// wired separately (see pkg/cmd/root) and are not part of this group definition.
package deviceregistration

import (
	cmdApprove "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceregistration/approve"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceregistration/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceregistration/get"
	cmdGetCredentials "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceregistration/getcredentials"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceregistration/list"
	cmdRegister "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceregistration/register"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
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
