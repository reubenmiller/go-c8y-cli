package cmd

import (
	"github.com/spf13/cobra"
)

type settingsCmd struct {
	*baseCmd
}

func newSettingsRootCmd() *settingsCmd {
	ccmd := &settingsCmd{}

	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Settings",
		Long:  `Settings commands`,
	}

	// Subcommands
	cmd.AddCommand(newListSettingsCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
