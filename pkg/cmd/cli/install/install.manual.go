package install

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/cli/safeexec"
	"github.com/mitchellh/go-homedir"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/shell"
	"github.com/spf13/cobra"
)

var ErrInstallFailed = errors.New("failed to install one or more profiles")
var validShells = shell.SupportedShells()

// Skip sh when installing all shells as it generally does not have a profile defined
var defaultShells = []string{shell.ShellBash, shell.ShellFish, shell.ShellPowershell, shell.ShellPwsh, shell.ShellZsh}

type CmdInstall struct {
	*subcommand.SubCommand

	shell   []string
	factory *cmdutil.Factory
}

func NewCmdInstall(f *cmdutil.Factory) *CmdInstall {
	ccmd := &CmdInstall{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install shell helpers",
		Long:  `Install shell helpers such as set-session etc.`,
		Example: heredoc.Doc(`
			$ c8y cli install
			Install helpers to all available shells

			$ c8y cli install --shell zsh
			Install zsh helpers

			$ c8y cli install --shell zsh,bash
			Install zsh and bash helpers
		`),
		RunE: ccmd.RunE,
	}

	cmd.Flags().StringSliceVar(&ccmd.shell, "shell", defaultShells, "Type of shell")

	cmd.SilenceUsage = true

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("shell", validShells...),
	)

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

type Shell struct {
	Name   string
	Binary string
}

func (n *CmdInstall) RunE(cmd *cobra.Command, args []string) error {
	shells := map[string]Shell{
		shell.ShellBash:       {Name: "bash", Binary: "bash"},
		shell.ShellZsh:        {Name: "zsh", Binary: "zsh"},
		shell.ShellPosixShell: {Name: "sh", Binary: "sh"},
		shell.ShellPwsh:       {Name: "pwsh", Binary: "pwsh"},
		shell.ShellPowershell: {Name: "powershell", Binary: "powershell"},
		shell.ShellFish:       {Name: "fish", Binary: "fish"},
	}

	var Errs []error
	reloadRequired := false

	// Detect shells
	detectedShells := make([]Shell, 0, len(n.shell))
	for _, name := range n.shell {
		shell, ok := shells[name]
		if !ok {
			slog.Warn("Skipping invalid shell type", "name", name)
			continue
		}

		slog.Debug("Checking shell", "name", shell.Name, "shell", shell.Binary)
		if _, err := safeexec.LookPath(shell.Binary); err == nil {
			detectedShells = append(detectedShells, shell)
		} else {
			slog.Debug("Shell was not found", "name", shell.Name, "shell", shell.Binary)
		}
	}

	if len(detectedShells) == 0 {
		// Add sh as a last resort, but since it does not have a commonly used
		// configuration file, don't display this warning all the time, otherwise it would be really
		// annoying to users as sh is installed on all Linux and macOS versions
		if _, err := safeexec.LookPath("sh"); err == nil {
			detectedShells = append(detectedShells, Shell{
				Name:   "sh",
				Binary: "sh",
			})
		} else {
			fmt.Fprint(n.factory.IOStreams.ErrOut, "Did not detect any valid shells\n")
			return nil
		}
	}

	// Install profiles in each shell that was found
	for _, shell := range detectedShells {
		changed, installErr := n.InstallProfile(shell.Name)
		if installErr != nil {
			Errs = append(Errs, installErr)
		}
		reloadRequired = reloadRequired || changed
	}

	if len(Errs) == 0 {
		if reloadRequired {
			fmt.Fprint(n.factory.IOStreams.ErrOut, "\nPlease reload your shell\n\n")
		}
		return nil
	}

	if len(Errs) == 1 {
		return Errs[0]
	}

	summaryErr := ErrInstallFailed
	for _, err := range Errs {
		summaryErr = fmt.Errorf("%w", err)
	}
	return summaryErr
}

func addExecutableBinaryToPowerShellPath(existingSnippet string) (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return existingSnippet, err
	}
	exeDir := filepath.Dir(exePath)

	snippet := fmt.Sprintf("$env:PATH += '%c%s'; ", filepath.ListSeparator, exeDir) + existingSnippet
	return snippet, nil
}

func (n *CmdInstall) InstallProfile(shellType string) (bool, error) {
	changed := false
	profilePath := ""
	profileSnippet := ""

	switch shellType {
	case shell.ShellPosixShell:
		// sh/ash uses a special env variable called 'ENV' which controls whether a profile is auto loaded or not
		profilePath = os.Getenv("ENV")
		profileSnippet = `eval "$(c8y cli profile --shell sh)"`
		if profilePath == "" {
			if n.factory.IOStreams != nil {
				fmt.Fprintf(
					n.factory.IOStreams.ErrOut,
					"%s sh is not using a config file (set via the 'ENV' env variable), so you will need to load the profile yourself using:\n\n  %s\n\n",
					n.factory.IOStreams.ColorScheme().FailureIcon(),
					profileSnippet,
				)
			}
			return false, cmderrors.NewSilentError()
		}

	case shell.ShellZsh:
		profilePath = "~/.zshrc"
		profileSnippet = "source <(c8y cli profile --shell zsh)"
	case shell.ShellBash:
		// use eval over source as process substitution was only added in bash >= v4
		profilePath = "~/.bashrc"
		profileSnippet = `eval "$(c8y cli profile --shell bash)"`
	case shell.ShellFish:
		profilePath = "~/.config/fish/config.fish"
		profileSnippet = "c8y cli profile --shell fish | source"
	case shell.ShellPowershell:
		profilePath = "~/Documents/WindowsPowerShell/Microsoft.PowerShell_profile.ps1"
		if snippet, err := addExecutableBinaryToPowerShellPath("c8y cli profile --shell powershell | Out-String | Invoke-Expression"); err == nil {
			profileSnippet = snippet
		} else {
			slog.Warn("Failed to detect binary path", "err", err)
		}
	case shell.ShellPwsh:
		profilePath = "~/.config/powershell/Microsoft.PowerShell_profile.ps1"
		if snippet, err := addExecutableBinaryToPowerShellPath("c8y cli profile --shell powershell | Out-String | Invoke-Expression"); err == nil {
			profileSnippet = snippet
		} else {
			slog.Warn("Failed to detect binary path", "err", err)
		}
	}

	if profilePath == "" {
		return changed, fmt.Errorf("profile path is empty")
	}

	expandedV, err := homedir.Expand(profilePath)
	if err != nil {
		return changed, err
	}

	// Create directory
	if err := os.MkdirAll(filepath.Dir(expandedV), 0755); err != nil {
		return changed, err
	}

	// Check if profile contains expected line
	appendToProfile := true

	if _, err := os.Stat(expandedV); !errors.Is(err, fs.ErrNotExist) {
		file, err := os.Open(expandedV)
		if err != nil {
			return changed, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), profileSnippet) {
				appendToProfile = false
				break
			}
		}
	} else {
		slog.Debug("Profile file does not exist", "path", expandedV)
	}

	cs := n.factory.IOStreams.ColorScheme()
	message := ""

	if appendToProfile {
		message = fmt.Sprintf("Added snippet to %s profile. path: %s", shellType, expandedV)
		slog.Debug("Adding snippet to profile", "path", expandedV)
		err = AppendToFile(expandedV, profileSnippet)
		if err != nil {
			return changed, err
		}
		changed = true
	} else {
		slog.Debug("Snippet already found in profile", "path", expandedV)
		message = fmt.Sprintf("Already added snippet to %s profile. path: %s", shellType, expandedV)
	}

	_, bErr := fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s %s\n", cs.SuccessIconWithColor(cs.Green), message)
	return changed, bErr
}

func AppendToFile(p string, text string) error {
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("%s\n", text))
	return err
}
