package profile

import (
	_ "embed"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/shell"
	"github.com/spf13/cobra"
)

//go:embed scripts/plugin.ps1
var scriptPowerShell string

//go:embed scripts/plugin.sh
var scriptBash string

//go:embed scripts/plugin.posix.sh
var scriptPosixShell string

//go:embed scripts/plugin.sh
var scriptZsh string

//go:embed scripts/plugin.auto.sh
var scriptLinuxAuto string

//go:embed scripts/plugin.fish
var scriptFish string

type CmdProfile struct {
	*subcommand.SubCommand

	shell   string
	factory *cmdutil.Factory
}

func NewCmdProfile(f *cmdutil.Factory) *CmdProfile {
	ccmd := &CmdProfile{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Print shell profile script",
		Long: heredoc.Doc(`
			Print the shell profile script which contains helpers such as set-session etc.

			This command can be used to load the associated shell helpers.
		`),
		Example: heredoc.Doc(`
		## zsh/bash/sh
			eval "$(c8y cli profile)"

		## zsh
			source <(c8y cli profile --shell zsh)
		
		## bash
			eval "$(c8y cli profile --shell bash)"
		
		## sh (posix)
			eval "$(c8y cli profile --shell sh)"

		## fish
			c8y cli profile --shell fish | source

		## PowerShell
			c8y cli profile --shell powershell | Out-String | Invoke-Expression
		`),
		RunE: ccmd.RunE,
	}

	cmd.Flags().StringVar(&ccmd.shell, "shell", "", "Type of shell")

	cmd.SilenceUsage = true

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("shell", shell.SupportedShells(shell.ShellAuto)...),
	)

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdProfile) RunE(cmd *cobra.Command, args []string) error {
	activeShell := n.shell
	if activeShell == "" {
		activeShell = shell.ShellPosixShell
	}
	var script *string

	// Note: for sh based shells, use a generic entrypoint
	// which is able to detect the exact shell type before running
	// the eval, as the process can't do it as it does not have access
	// to the special shell variables as they are not exposed as env variables
	// so a minimal source is still required to then detect the correct variant,
	// and then the real shell type (with a prefixed "__{shell}") is used after
	// the real shell type has been determined.
	// This does not work for fish, or powershell as the syntax is too different
	switch activeShell {
	case shell.ShellZsh, shell.ShellBash, shell.ShellPosixShell, shell.ShellAuto:
		// generic entry point which will find the correct shell
		script = &scriptLinuxAuto
	case "__zsh":
		script = &scriptZsh
	case "__bash":
		script = &scriptBash
	case "__sh":
		script = &scriptPosixShell
	case shell.ShellFish:
		script = &scriptFish
	case shell.ShellPowershell, shell.ShellPwsh:
		script = &scriptPowerShell
	}

	if script != nil {
		n.GetCommand().Printf("%s\n", *script)
	}

	return nil
}
