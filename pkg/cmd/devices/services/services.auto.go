package services

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/services/create"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/services/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/services/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdServices struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdServices {
	ccmd := &SubCmdServices{}

	cmd := &cobra.Command{
		Use:   "services",
		Short: "Cumulocity device services",
		Long:  `Managed device services (introduced in 10.14)`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
