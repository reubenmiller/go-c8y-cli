package plugins

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/ui/applications/plugins/delete"
	cmdInstall "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/ui/applications/plugins/install"
	cmdReplace "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/ui/applications/plugins/replace"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/ui/applications/plugins/update"
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
		Short: "Cumulocity IoT UI Application plugin management",
		Long:  `Manage the plugins which are installed in a UI application`,
	}

	// Subcommands
	cmd.AddCommand(cmdInstall.NewCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewCmd(f).GetCommand())
	cmd.AddCommand(cmdReplace.NewCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewCmdUpdate(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
