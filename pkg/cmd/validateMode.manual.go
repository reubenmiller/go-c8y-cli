package cmd

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/spf13/cobra"
)

func getValidationError(mode string, setting string) error {
	return fmt.Errorf("%s mode is disabled. %s commands are disabled unless '%s' is set to 'true' in your session settings", mode, mode, setting)
}

func validateCreateMode(cmd *cobra.Command, args []string) error {
	if !cliConfig.AllowModeCreate() {
		return getValidationError("create", config.SettingsModeEnableCreate)
	}
	return nil
}

func validateUpdateMode(cmd *cobra.Command, args []string) error {
	if !cliConfig.AllowModeUpdate() {
		return getValidationError("update", config.SettingsModeEnableUpdate)
	}
	return nil
}

func validateDeleteMode(cmd *cobra.Command, args []string) error {
	if !cliConfig.AllowModeDelete() {
		return getValidationError("delete", config.SettingsModeEnableDelete)
	}
	return nil
}
