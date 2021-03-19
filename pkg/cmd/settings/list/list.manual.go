package cmd

import (
	"encoding/json"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/spf13/cobra"
)

type CmdList struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdList(f *cmdutil.Factory) *CmdList {
	ccmd := &CmdList{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show the current settings",
		Long:  `Show the current settings which are being used by the cli tool`,
		Example: heredoc.Doc(`
$ c8y settings list
Show the all settings (as json)

$ c8y settings list --select "activityLog" --flatten
Show active log settings in a flattened json format
		`),
		RunE: ccmd.listSettings,
	}

	cmd.SilenceUsage = true
	cmdutil.DisableAuthCheck(cmd)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdList) listSettings(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	activitylog, err := n.factory.ActivityLogger()
	if err != nil {
		return err
	}
	var responseText []byte

	// settings := viper.GetViper().AllSettings()
	settings := cfg.AllSettings()

	allSettings := mapbuilder.NewInitializedMapBuilder()
	allSettings.ApplyMap(settings)

	// add additional settings
	err = allSettings.Set("settings.session.home", cfg.GetSessionHomeDir())
	if err != nil {
		cfg.Logger.Warnf("Could not get home session directory. %s", err)
	}

	if activitylog != nil {
		err := allSettings.Set("settings.activitylog.currentPath", activitylog.GetPath())
		if err != nil {
			cfg.Logger.Warnf("Could not get activity logger path. %s", err)
		}
	}

	responseText, err = json.Marshal(allSettings)
	if err != nil {
		return cmderrors.NewUserError("Settings error. ", err)
	}

	return n.factory.WriteJSONToConsole(cfg, cmd, "settings", responseText)
}
