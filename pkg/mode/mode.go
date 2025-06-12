package mode

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

func getValidationError(mode string, recommendedModes []string) error {
	return fmt.Errorf("%s mode is disabled. %s commands are disabled unless the mode is one of: %q", mode, mode, recommendedModes)
}

func ValidateCreateMode(cfg *config.Config) error {
	if !cfg.AllowModeCreate() {
		modes := []string{
			config.SessionModeDev.String(),
			config.SessionModeQual.String(),
			config.SessionModeCI.String(),
		}
		return getValidationError("create", modes)
	}
	return nil
}

func ValidateUpdateMode(cfg *config.Config) error {
	if !cfg.AllowModeUpdate() {
		modes := []string{
			config.SessionModeDev.String(),
			config.SessionModeQual.String(),
			config.SessionModeCI.String(),
		}
		return getValidationError("update", modes)
	}
	return nil
}

func ValidateDeleteMode(cfg *config.Config) error {
	if !cfg.AllowModeDelete() {
		modes := []string{
			config.SessionModeDev.String(),
			config.SessionModeCI.String(),
		}
		return getValidationError("delete", modes)
	}
	return nil
}
