package patches

import (
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/versions/patches/delete"
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
		Short: "Cumulocity firmware version patch management",
		Long:  `REST endpoint to interact with Cumulocity firmware versions patches`,
	}

	// Subcommands
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
