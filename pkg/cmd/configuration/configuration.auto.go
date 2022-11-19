package configuration

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/configuration/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/configuration/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/configuration/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/configuration/list"
	cmdSend "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/configuration/send"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/configuration/update"
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
		Short: "Cumulocity configuration repository management",
		Long:  `Configuration management to create/list/delete configurations`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdSend.NewSendCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
