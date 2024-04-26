package plugins

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/ui/plugins/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/ui/plugins/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/ui/plugins/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/ui/plugins/update"
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
		Short: "Cumulocity IoT UI Plugins",
		Long:  `Managed UI Plugins`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
