package cmd

import (
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/settings/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/settings/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"

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
	cmd.AddCommand(cmdList.NewCmdList(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewCmdUpdate(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
