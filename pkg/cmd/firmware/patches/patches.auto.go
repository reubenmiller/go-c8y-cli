package patches

import (
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/patches/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/patches/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/patches/list"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdPatches struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdPatches {
	ccmd := &SubCmdPatches{}

	cmd := &cobra.Command{
		Use:   "patches",
		Short: "Cumulocity firmware patch management",
		Long:  `REST endpoint to interact with Cumulocity firmware patches`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
