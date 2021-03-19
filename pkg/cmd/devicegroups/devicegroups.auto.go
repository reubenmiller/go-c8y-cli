package devicegroups

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicegroups/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicegroups/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicegroups/get"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicegroups/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdDevicegroups struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdDevicegroups {
	ccmd := &SubCmdDevicegroups{}

	cmd := &cobra.Command{
		Use:   "devicegroups",
		Short: "Cumulocity device groups",
		Long:  `REST endpoint to interact with Cumulocity device groups`,
	}

	// Subcommands
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
