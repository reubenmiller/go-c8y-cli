package utils

import "os"

func IsDebugEnabled() (bool, string) {
	debugValue, isDebugSet := os.LookupEnv("GH_DEBUG")
	legacyDebugValue := os.Getenv("DEBUG")

	if !isDebugSet {
		switch legacyDebugValue {
		case "true", "1", "yes", "api":
			return true, legacyDebugValue
		default:
			return false, legacyDebugValue
		}
	}

	switch debugValue {
	case "false", "0", "no", "":
		return false, debugValue
	default:
		return true, debugValue
	}
}
