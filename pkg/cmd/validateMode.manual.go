package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func getValidationError(mode string, setting string) error {
	return fmt.Errorf("%s mode is disabled. %s commands are disabled unless '%s' is set to 'true' in your session settings", mode, mode, setting)
}

func validateBatchCreateMode(cmd *cobra.Command, args []string) error {
	if !(globalModeEnableBatch || globalModeEnableCreate || globalCIMode) {
		return getValidationError("batch", SettingsModeEnableBatch)
	}
	return nil
}

func validateBatchUpdateMode(cmd *cobra.Command, args []string) error {
	if !(globalModeEnableBatch || globalModeEnableUpdate || globalCIMode) {
		return getValidationError("batch", SettingsModeEnableBatch)
	}
	return nil
}

func validateBatchDeleteMode(cmd *cobra.Command, args []string) error {
	if !(globalModeEnableBatch || globalModeEnableDelete || globalCIMode) {
		return getValidationError("batch", SettingsModeEnableBatch)
	}
	return nil
}

func validateCreateMode(cmd *cobra.Command, args []string) error {
	if !(globalModeEnableCreate || globalCIMode) {
		return getValidationError("create", SettingsModeEnableCreate)
	}
	return nil
}

func validateUpdateMode(cmd *cobra.Command, args []string) error {
	if !(globalModeEnableUpdate || globalCIMode) {
		return getValidationError("update", SettingsModeEnableUpdate)
	}
	return nil
}

func validateDeleteMode(cmd *cobra.Command, args []string) error {
	if !(globalModeEnableDelete || globalCIMode) {
		return getValidationError("delete", SettingsModeEnableDelete)
	}
	return nil
}
