package connect

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/connect/ssh"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdConnect struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdConnect {
	ccmd := &SubCmdConnect{}

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to a device",
		Long:  `Connect to a device using the protocol defined in an existing configuration`,
	}

	// Subcommands
	cmd.AddCommand(ssh.NewCmdSSH(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
