package util

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	cmdRepeat "github.com/reubenmiller/go-c8y-cli/pkg/cmd/util/repeat"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
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
	cmd.AddCommand(cmdRepeat.NewCmdRepeat(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
