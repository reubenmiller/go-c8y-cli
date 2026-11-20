package util

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdUtil struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdUtil {
	ccmd := &SubCmdUtil{}

	cmd := &cobra.Command{
		Use:   "util",
		Short: "General utilities",
		Long:  `General utilities for combining with other c8y commands`,
	}

	// Subcommands
	cmd.AddCommand(NewCmdRepeat(f).GetCommand())
	cmd.AddCommand(NewCmdRepeatFile(f).GetCommand())
	cmd.AddCommand(NewCmdShow(f).GetCommand())
	cmd.AddCommand(NewCmdFromCsv(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
