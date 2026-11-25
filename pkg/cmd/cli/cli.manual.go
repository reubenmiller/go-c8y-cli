package cli

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdCLI struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdCLI {
	ccmd := &SubCmdCLI{}

	cmd := &cobra.Command{
		Use:   "cli",
		Short: "cli management commands",
		Long:  `Commands used to managed go-c8y-cli`,
	}

	// Subcommands
	cmd.AddCommand(NewCmdInstall(f).GetCommand())
	cmd.AddCommand(NewCmdProfile(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
