// Package configurations groups the remote access configuration subcommands
// (list/get/update/delete and the create-passthrough/webssh/vnc/telnet
// variants), all hand-written v2 c8ystream commands backed by the go-c8y v2
// RemoteAccess.Configurations service.
package configurations

import (
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
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewCreatePassthroughCmd(f).GetCommand())
	cmd.AddCommand(NewCreateWebsshCmd(f).GetCommand())
	cmd.AddCommand(NewCreateVncCmd(f).GetCommand())
	cmd.AddCommand(NewCreateTelnetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
