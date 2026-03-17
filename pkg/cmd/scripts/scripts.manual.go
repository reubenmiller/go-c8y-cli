package scripts

import (
	cmdFormat "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/scripts/format"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdScripts struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdScripts {
	ccmd := &SubCmdScripts{}

	cmd := &cobra.Command{
		Use:   "scripts",
		Short: "Script utilities",
		Long:  `Utilities for processing and formatting shell scripts`,
	}

	// Subcommands
	cmd.AddCommand(cmdFormat.NewCmdFormat(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
