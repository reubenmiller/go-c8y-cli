package shell

import "os"

// DetectShell detect the shell type, i.e. fish, bash, zsh or powershell
func DetectShell(defaultValue string) string {
	if os.Getenv("FISH_VERSION") != "" {
		return "fish"
	}
	if os.Getenv("BASH_VERSION") != "" {
		return "bash"
	}
	if os.Getenv("ZSH_VERSION") != "" {
		return "zsh"
	}
	if os.Getenv("PSModulePath") != "" {
		return "powershell"
	}
	return defaultValue
}
