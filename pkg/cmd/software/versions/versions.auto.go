package versions

import (
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/software/versions/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/software/versions/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/software/versions/list"
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
		Short: "Cumulocity software version management",
		Long:  `REST endpoint to interact with Cumulocity software versions`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
