package settings

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"

	"github.com/spf13/cobra"
)

type SubCmdSettings struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdSettings {
	ccmd := &SubCmdSettings{}

	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Settings",
		Long:  `Settings commands`,
	}

	// Subcommands
	cmd.AddCommand(NewCmdList(f).GetCommand())
	cmd.AddCommand(NewCmdUpdate(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
