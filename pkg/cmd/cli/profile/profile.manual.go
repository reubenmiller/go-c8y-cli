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

//go:embed scripts/plugin.sh
var scriptZsh string

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
		## zsh
			source <(c8y cli profile --shell zsh)
		
		## bash
			source <(c8y cli profile --shell bash)

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
		completion.WithValidateSet("shell", "bash", "zsh", "powershell", "fish"),
	)

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdProfile) RunE(cmd *cobra.Command, args []string) error {
	activeShell := n.shell
	if n.shell == "" {
		activeShell = shell.DetectShell("zsh")
	}
	var script *string

	switch activeShell {
	case "zsh":
		script = &scriptZsh
	case "bash":
		script = &scriptBash
	case "fish":
		script = &scriptFish
	case "powershell":
		script = &scriptPowerShell
	}

	if script != nil {
		n.GetCommand().Printf("%s\n", *script)
	}

	return nil
}
