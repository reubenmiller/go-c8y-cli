package configurations

import (
	cmdCreatePassthrough "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configurations/create_passthrough"
	cmdCreateTelnet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configurations/create_telnet"
	cmdCreateVnc "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configurations/create_vnc"
	cmdCreateWebssh "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configurations/create_webssh"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configurations/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configurations/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configurations/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configurations/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdConfigurations struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdConfigurations {
	ccmd := &SubCmdConfigurations{}

	cmd := &cobra.Command{
		Use:   "configurations",
		Short: "Manage remote access configurations",
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
