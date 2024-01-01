package cli

import (
	cmdInstall "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/cli/install"
	cmdProfile "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/cli/profile"
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
	cmd.AddCommand(cmdInstall.NewCmdInstall(f).GetCommand())
	cmd.AddCommand(cmdProfile.NewCmdProfile(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
