package versions

import (
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/firmware/versions/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/firmware/versions/get"
	cmdInstall "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/firmware/versions/install"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/firmware/versions/list"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdVersions struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdVersions {
	ccmd := &SubCmdVersions{}

	cmd := &cobra.Command{
		Use:   "versions",
		Short: "Cumulocity firmware version management",
		Long:  `Firmware version management to create/list/delete versions`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdInstall.NewInstallCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
