package applications

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	cmdPlugins "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/ui/applications/plugins"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdExtensions struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdExtensions {
	ccmd := &SubCmdExtensions{}

	cmd := &cobra.Command{
		Use:   "applications",
		Short: "Cumulocity IoT UI Applications",
		Long:  `Managed UI Applications`,
	}

	// Subcommands
	cmd.AddCommand(cmdPlugins.NewSubCommand(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
