package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

type listSettingsCmd struct {
	*baseCmd
}

func newListSettingsCmd() *listSettingsCmd {
	ccmd := &listSettingsCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show the current settings",
		Long:  `Show the current settings which are being used by the cli tool`,
		Example: `
$ c8y settings list

Show the active cli settings
		`,
		RunE: ccmd.listSettings,
	}

	cmd.SilenceUsage = true

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *listSettingsCmd) listSettings(cmd *cobra.Command, args []string) error {

	settings := map[string]interface{}{}

	settings["session.home"] = getSessionHomeDir()

	settings[SettingsTemplatePath] = globalFlagTemplatePath
	settings[SettingsDefaultPageSize] = globalFlagPageSize
	settings[SettingsIncludeAllPageSize] = globalFlagIncludeAllPageSize
	settings[SettingsIncludeAllDelayMS] = globalFlagIncludeAllDelayMS

	settings[SettingsModeEnableCreate] = globalModeEnableCreate
	settings[SettingsModeEnableUpdate] = globalModeEnableUpdate
	settings[SettingsModeEnableDelete] = globalModeEnableDelete
	settings[SettingsEncryptionEnabled] = cliConfig.IsEncryptionEnabled()
	settings[SettingsModeCI] = globalCIMode

	responseText, err := json.Marshal(settings)

	if err != nil {
		return newUserError("Settings error. ", err)
	}

	if globalFlagCompact {
		fmt.Printf("%s\n", bytes.TrimSpace(responseText))
	} else {
		fmt.Printf("%s", pretty.Pretty(bytes.TrimSpace(responseText)))
	}
	return nil
}
