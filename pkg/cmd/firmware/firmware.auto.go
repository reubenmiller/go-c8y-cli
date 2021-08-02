package firmware

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdFirmware struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdFirmware {
	ccmd := &SubCmdFirmware{}

	cmd := &cobra.Command{
		Use:   "firmware",
		Short: "Cumulocity firmware management",
		Long:  `REST endpoint to interact with Cumulocity managed objects`,
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
