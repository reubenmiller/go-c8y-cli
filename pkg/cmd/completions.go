package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// CompletionsCmd shell completions
type CompletionsCmd struct {
	*baseCmd
}

// NewCompletionsCmd creates a new shell completions command to generate completions files in variables shell languages
func NewCompletionsCmd() *CompletionsCmd {
	ccmd := &CompletionsCmd{}

	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:
	
	Bash:
	
	  $ source <(c8y completion bash)
	
	  # To load completions for each session, execute once:
	  # Linux:
	  $ c8y completion bash > /etc/bash_completion.d/c8y
	  # macOS:
	  $ c8y completion bash > /usr/local/etc/bash_completion.d/c8y
	
	Zsh:
	
	  # If shell completion is not already enabled in your environment,
	  # you will need to enable it.  You can execute the following once:
	
	  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
	
	  # To load completions for each session, execute once:
	  $ c8y completion zsh > "${fpath[1]}/_c8y"
	
	  # You will need to start a new shell for this setup to take effect.
	
	fish:
	
	  $ set -x LANG C.UTF-8
	  $ c8y completion fish | source
	
	  # To load completions for each session, execute once:
	  $ c8y completion fish > ~/.config/fish/completions/c8y.fish
	
	PowerShell:
	
	  PS> c8y completion powershell | Out-String | Invoke-Expression
	
	  # To load completions for every new session, run:
	  PS> c8y completion powershell > c8y.ps1
	  # and source this file from your PowerShell profile.
	`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
