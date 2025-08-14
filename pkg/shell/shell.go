package shell

import (
	"os"
	"slices"
)

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

var (
	ShellFish       = "fish"
	ShellBash       = "bash"
	ShellZsh        = "zsh"
	ShellPowershell = "powershell"
	ShellPwsh       = "pwsh"
	ShellPosixShell = "sh"
	ShellAuto       = "auto"
)

func SupportedShells(additional ...string) []string {
	values := []string{ShellFish, ShellBash, ShellZsh, ShellPosixShell, ShellPowershell}
	for _, v := range additional {
		if !slices.Contains(values, v) {
			values = append(values, v)
		}
	}
	return values
}

func SupportTabCompletionShells(additional ...string) []string {
	values := []string{ShellFish, ShellBash, ShellZsh, ShellPowershell, ShellPwsh}
	for _, v := range additional {
		if !slices.Contains(values, v) {
			values = append(values, v)
		}
	}
	return values
}
