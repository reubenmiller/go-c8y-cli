package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type settingsCmd struct {
	*subcommand.SubCommand
}

func NewSettingsRootCmd() *settingsCmd {
	ccmd := &settingsCmd{}

	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Settings",
		Long:  `Settings commands`,
	}

	// Subcommands
	cmd.AddCommand(newListSettingsCmd().GetCommand())
	cmd.AddCommand(NewUpdateSettingsCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
