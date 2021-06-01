package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// SetMode set the session mode to control which commands are enabled or not
func SetMode(v *viper.Viper, mode string) error {
	switch strings.ToLower(mode) {
	case "prod":
		v.Set(SettingsModeEnableCreate, false)
		v.Set(SettingsModeEnableUpdate, false)
		v.Set(SettingsModeEnableDelete, false)
		v.Set(SettingsModeCI, false)

	case "qual":
		v.Set(SettingsModeEnableCreate, true)
		v.Set(SettingsModeEnableUpdate, true)
		v.Set(SettingsModeEnableDelete, false)
		v.Set(SettingsModeCI, false)

	case "dev":
		v.Set(SettingsModeEnableCreate, true)
		v.Set(SettingsModeEnableUpdate, true)
		v.Set(SettingsModeEnableDelete, true)
		v.Set(SettingsModeCI, false)

	case "ci":
		v.Set(SettingsModeCI, true)
	default:
		return fmt.Errorf("unsupported mode. %s. Supported modes are [dev, qual, prod, ci]", mode)
	}
	return nil
}
