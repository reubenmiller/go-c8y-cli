package cmd

import (
	"encoding/json"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
Show the all settings (as json)

$ c8y settings list --select "activityLog" --flatten
Show active log settings in a flattened json format
		`,
		RunE: ccmd.listSettings,
	}

	cmd.SilenceUsage = true

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *listSettingsCmd) listSettings(cmd *cobra.Command, args []string) error {
	var responseText []byte
	var err error

	settings := viper.GetViper().AllSettings()

	allSettings := mapbuilder.NewInitializedMapBuilder()
	allSettings.ApplyMap(settings)

	// add additional settings
	err = allSettings.Set("settings.session.home", getSessionHomeDir())
	if err != nil {
		Logger.Warnf("Could not get home session directory. %s", err)
	}

	if activityLogger != nil {
		err := allSettings.Set("settings.activitylog.currentPath", activityLogger.GetPath())
		if err != nil {
			Logger.Warnf("Could not get activity logger path. %s", err)
		}
	}

	responseText, err = json.Marshal(allSettings)
	if err != nil {
		return cmderrors.NewUserError("Settings error. ", err)
	}

	return WriteJSONToConsole(cmd, "settings", responseText)
}
