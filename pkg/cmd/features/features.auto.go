package features

import (
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/features/delete"
	cmdDisable "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/features/disable"
	cmdEnable "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/features/enable"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/features/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/features/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/features/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdFeatures struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdFeatures {
	ccmd := &SubCmdFeatures{}

	cmd := &cobra.Command{
		Use:   "features",
		Short: "Cumulocity tenant features",
		Long:  `Managed feature toggles`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdEnable.NewEnableCmd(f).GetCommand())
	cmd.AddCommand(cmdDisable.NewDisableCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
