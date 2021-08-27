package versions

import (
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/versions/delete"
	cmdDownload "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/versions/download"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/versions/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/versions/list"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
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
		Long:  `REST endpoint to interact with Cumulocity firmware versions`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdDownload.NewDownloadCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}