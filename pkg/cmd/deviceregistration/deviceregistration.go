// Package deviceregistration wires the `c8y deviceregistration` command and its
// spec-generated subcommands. All of them are hand-written v2 c8ystream commands
// (backed by the go-c8y v2 Devices.Registration service), so this group command
// is itself hand-written rather than generated. The bulk register-* commands are
// wired separately (see pkg/cmd/root) and are not part of this group definition.
package deviceregistration

import (
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
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewRegisterCmd(f).GetCommand())
	cmd.AddCommand(NewApproveCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewGetCredentialsCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
