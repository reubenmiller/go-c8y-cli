package remoteaccess

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/configuration"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/connect"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/server"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdRemoteAccess struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdRemoteAccess {
	ccmd := &SubCmdRemoteAccess{}

	cmd := &cobra.Command{
		Use:   "remoteaccess",
		Short: "Cumulocity remoteaccess management",
		Long:  `Cumulocity remoteaccess management`,
	}

	// Subcommands
	cmd.AddCommand(connect.NewSubCommand(f).GetCommand())
	cmd.AddCommand(server.NewCmdServer(f).GetCommand())
	cmd.AddCommand(configuration.NewSubCommand(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
