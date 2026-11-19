// Package deviceprofiles wires the `c8y deviceprofiles` command and its
// subcommands. All subcommands are hand-written v2 c8ystream commands (backed by
// the go-c8y v2 DeviceProfiles service), so this group command is itself
// hand-written rather than generated.
package deviceprofiles

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceprofiles/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceprofiles/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceprofiles/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceprofiles/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceprofiles/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdDeviceprofiles struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdDeviceprofiles {
	ccmd := &SubCmdDeviceprofiles{}

	cmd := &cobra.Command{
		Use:   "deviceprofiles",
		Short: "Cumulocity device profile management",
		Long:  `Commands to managed Cumulocity device profiles`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
