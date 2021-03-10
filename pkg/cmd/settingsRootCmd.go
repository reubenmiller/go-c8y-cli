package cmd

import (
	"github.com/spf13/cobra"
)

type settingsCmd struct {
	*baseCmd
}

func NewSettingsRootCmd() *settingsCmd {
	ccmd := &settingsCmd{}

	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Settings",
		Long:  `Settings commands`,
	}

	// Subcommands
	cmd.AddCommand(newListSettingsCmd().getCommand())
	cmd.AddCommand(NewUpdateSettingsCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
