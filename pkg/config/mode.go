package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// SetMode set the session mode to control which commands are enabled
func SetMode(v *viper.Viper, mode SessionMode) error {
	v.Set(SettingsMode, mode.String())
	if mode != SessionModeCI {
		v.Set(SettingsModeCI, false)
	}
	return nil
}

// SessionMode session mode to control which commands are enabled
type SessionMode int

const (
	// SessionModeProduction production / read-only mode
	SessionModeProduction SessionMode = iota

	// SessionModeQual QA / create / update mode
	SessionModeQual

	// SessionModeDev QA / create / update / delete mode
	SessionModeDev

	// SessionModeCI CI mode where everything is allowed
	SessionModeCI
)

func (f SessionMode) String() string {
	values := map[SessionMode]string{
		SessionModeProduction: "prod",
		SessionModeQual:       "qual",
		SessionModeDev:        "dev",
		SessionModeCI:         "ci",
	}

	if v, ok := values[f]; ok {
		return v
	}
	return ""
}

func (f SessionMode) FromString(name string, ci bool) SessionMode {
	if ci {
		return SessionModeCI
	}
	values := map[string]SessionMode{
		"prod": SessionModeProduction,
		"qual": SessionModeQual,
		"dev":  SessionModeDev,
		"ci":   SessionModeCI,
	}

	if v, ok := values[strings.ToLower(name)]; ok {
		return v
	}
	return f
}

func (f SessionMode) CanCreate() bool {
	return f == SessionModeCI || f == SessionModeDev || f == SessionModeQual
}

func (f SessionMode) CanUpdate() bool {
	return f == SessionModeCI || f == SessionModeDev || f == SessionModeQual
}

func (f SessionMode) CanDelete() bool {
	return f == SessionModeCI || f == SessionModeDev
}

func GetSessionModeCompletionHelp() []string {
	return []string{
		fmt.Sprintf("%s\tProduction mode (read only)", SessionModeProduction.String()),
		fmt.Sprintf("%s\tQA mode (delete disabled)", SessionModeQual.String()),
		fmt.Sprintf("%s\tDevelopment mode (no restrictions)", SessionModeDev.String()),
		fmt.Sprintf("%s\tCI mode (no restrictions)", SessionModeCI.String()),
	}
}
