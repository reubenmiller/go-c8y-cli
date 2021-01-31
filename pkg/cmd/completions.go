package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

type completionsCmd struct {
	*baseCmd
}

func NewCompletionsCmd() *completionsCmd {
	ccmd := &completionsCmd{}

	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generates completion scripts",
		Long: `To load completion run

. <(bitbucket completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(bitbucket completion)
`,
	}

	// Subcommands
	cmd.AddCommand(newBashCompletionCmd().getCommand())
	cmd.AddCommand(newZshCompletionCmd().getCommand())
	cmd.AddCommand(newPowershellCompletionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

type bashCompletionCmd struct {
	*baseCmd
}

func newBashCompletionCmd() *bashCompletionCmd {
	ccmd := &bashCompletionCmd{}

	cmd := &cobra.Command{
		Use:   "bash",
		Short: "Generates bash completion scripts",
		Long: `To load completion run

. <(bitbucket completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(bitbucket completion)
`,
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
	}

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

type zshCompletionCmd struct {
	*baseCmd
}

func newZshCompletionCmd() *zshCompletionCmd {
	ccmd := &zshCompletionCmd{}

	cmd := &cobra.Command{
		Use:   "zsh",
		Short: "Generates zsh completion scripts",
		Long: `To load completion run

. <(bitbucket completion)

To configure your zsh shell to load completions for each session add to your bashrc

# ~/.zshrc or ~/.profile
. <(bitbucket completion)
`,
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenZshCompletion(os.Stdout)
		},
	}

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

type powershellCompletionCmd struct {
	*baseCmd
}

func newPowershellCompletionCmd() *powershellCompletionCmd {
	ccmd := &powershellCompletionCmd{}

	cmd := &cobra.Command{
		Use:   "powershell",
		Short: "Generates powershell completion scripts",
		Long: `To load completion run

. <(bitbucket completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(bitbucket completion)
`,
		Run: func(cmd *cobra.Command, args []string) {
			newPowershellCompletionHelper(&rootCmd.Command).GenPowerShellCompletion(os.Stdout)
		},
	}

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
