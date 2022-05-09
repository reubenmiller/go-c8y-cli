package user

import (
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/user/get"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/user/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdUser struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdUser {
	ccmd := &SubCmdUser{}

	cmd := &cobra.Command{
		Use:   "user",
		Short: "Cumulocity device user management",
		Long:  `Managed the device user related to a device`,
	}

	// Subcommands
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
