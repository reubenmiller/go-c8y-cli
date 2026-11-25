package plugins

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdPlugins struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdPlugins {
	ccmd := &SubCmdPlugins{}

	cmd := &cobra.Command{
		Use:   "plugins",
		Short: "Cumulocity UI Application plugin management",
		Long:  `Manage the plugins which are installed in a UI application`,
	}

	// Subcommands
	cmd.AddCommand(InstallNewCmd(f).GetCommand())
	cmd.AddCommand(DeleteNewCmd(f).GetCommand())
	cmd.AddCommand(ReplaceNewCmd(f).GetCommand())
	cmd.AddCommand(NewCmdUpdate(f).GetCommand())
	cmd.AddCommand(ListNewCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
