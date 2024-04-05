package extensions

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/ui/extensions/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/ui/extensions/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/ui/extensions/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/ui/extensions/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdExtensions struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdExtensions {
	ccmd := &SubCmdExtensions{}

	cmd := &cobra.Command{
		Use:   "extensions",
		Short: "Cumulocity IoT UI Extensions",
		Long:  `Managed UI extensions`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
