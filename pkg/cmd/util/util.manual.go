package util

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	cmdRepeat "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/util/repeat"
	cmdRepeatCsvFile "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/util/repeatcsv"
	cmdRepeatFile "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/util/repeatfile"
	cmdShow "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/util/show"
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
	cmd.AddCommand(cmdRepeat.NewCmdRepeat(f).GetCommand())
	cmd.AddCommand(cmdRepeatFile.NewCmdRepeatFile(f).GetCommand())
	cmd.AddCommand(cmdShow.NewCmdShow(f).GetCommand())
	cmd.AddCommand(cmdRepeatCsvFile.NewCmdFromCsv(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
