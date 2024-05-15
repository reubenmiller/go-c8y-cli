package configuration

import (
	cmdCreatePassthrough "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configuration/create_passthrough"
	cmdCreateTelnet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configuration/create_telnet"
	cmdCreateVnc "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configuration/create_vnc"
	cmdCreateWebssh "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configuration/create_webssh"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configuration/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configuration/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configuration/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configuration/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdConfiguration struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdConfiguration {
	ccmd := &SubCmdConfiguration{}

	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Manage Cloud Remote Access configuration",
		Long:  `Cloud Remote Access configuration management`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdCreatePassthrough.NewCreatePassthroughCmd(f).GetCommand())
	cmd.AddCommand(cmdCreateWebssh.NewCreateWebsshCmd(f).GetCommand())
	cmd.AddCommand(cmdCreateVnc.NewCreateVncCmd(f).GetCommand())
	cmd.AddCommand(cmdCreateTelnet.NewCreateTelnetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
