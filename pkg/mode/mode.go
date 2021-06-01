package mode

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/pkg/config"
)

func getValidationError(mode string, setting string) error {
	return fmt.Errorf("%s mode is disabled. %s commands are disabled unless '%s' is set to 'true' in your session settings", mode, mode, setting)
}

func ValidateCreateMode(cfg *config.Config) error {
	if !cfg.AllowModeCreate() {
		return getValidationError("create", config.SettingsModeEnableCreate)
	}
	return nil
}

func ValidateUpdateMode(cfg *config.Config) error {
	if !cfg.AllowModeUpdate() {
		return getValidationError("update", config.SettingsModeEnableUpdate)
	}
	return nil
}

func ValidateDeleteMode(cfg *config.Config) error {
	if !cfg.AllowModeDelete() {
		return getValidationError("delete", config.SettingsModeEnableDelete)
	}
	return nil
}
