package extensions

import (
	"os"
	"runtime"
	"strings"
)

const (
	// ContextVersion is the version of the runtime context (environment variables) which is
	// passed to extensions. Increase it when variables are added or changed, together with
	// extsdk.ContextVersion
	ContextVersion = "1"

	// Environment variables passed to extensions
	EnvExtensionName    = "C8Y_EXTENSION_NAME"
	EnvExtensionDir     = "C8Y_EXTENSION_DIR"
	EnvContextVersion   = "C8Y_EXTENSION_CONTEXT_VERSION"
	EnvBin              = "C8Y_BIN"
	EnvSettingsDryRun   = "C8Y_SETTINGS_DEFAULTS_DRY"
	EnvSettingsMode     = "C8Y_SETTINGS_MODE"
	EnvSettingsInsecure = "C8Y_SETTINGS_DEFAULTS_INSECURE"

	// EnvNullInput is "true" when the extension should not read stdin, e.g. to detect piped
	// input (set by --nullInput/-n). c8y commands don't use it, so the pipes within a script
	// still work, but it is inherited by the extension commands which the extension runs
	EnvNullInput = "C8Y_EXTENSION_NULL_INPUT"
)

// Variables returns the extension variables (name, folder, path to c8y, and context version).
// fallbackBin is used for C8Y_BIN if the path of the running executable can't be determined
func Variables(ext Extension, fallbackBin string) map[string]string {
	vars := map[string]string{
		EnvExtensionName:  ext.Name(),
		EnvExtensionDir:   ext.Path(),
		EnvContextVersion: ContextVersion,
	}
	if exe, err := os.Executable(); err == nil {
		vars[EnvBin] = exe
	} else if fallbackBin != "" {
		vars[EnvBin] = fallbackBin
	}
	return vars
}

// MergeEnv returns base ("KEY=value" items) without the variables in set and unset, followed by
// the variables in set. Names are compared case-insensitively on Windows
func MergeEnv(base []string, set map[string]string, unset map[string]bool) []string {
	sameKey := func(a, b string) bool {
		if runtime.GOOS == "windows" {
			return strings.EqualFold(a, b)
		}
		return a == b
	}
	isReplaced := func(name string) bool {
		for key := range set {
			if sameKey(key, name) {
				return true
			}
		}
		for key := range unset {
			if sameKey(key, name) {
				return true
			}
		}
		return false
	}

	env := make([]string, 0, len(base)+len(set))
	for _, item := range base {
		name, _, _ := strings.Cut(item, "=")
		if isReplaced(name) {
			continue
		}
		env = append(env, item)
	}
	for key, value := range set {
		env = append(env, key+"="+value)
	}
	return env
}
