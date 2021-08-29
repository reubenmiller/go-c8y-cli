package configuration

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/configuration/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/configuration/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/configuration/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/configuration/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/configuration/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
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
		Long:  `Managed configurations in the configuration repository`,
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
