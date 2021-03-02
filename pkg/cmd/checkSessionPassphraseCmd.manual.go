package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/reubenmiller/go-c8y-cli/pkg/prompt"
	"github.com/spf13/cobra"
)

type passphraseExchange struct {
	passphrase string ``
}

type checkSessionPassphraseCmd struct {
	OutputJSON         bool
	OutputEnvVariables bool
	prompter           *prompt.Prompt
	*baseCmd
}

func newCheckSessionPassphraseCmd() *checkSessionPassphraseCmd {
	ccmd := &checkSessionPassphraseCmd{}
	ccmd.prompter = prompt.NewPrompt(Logger)

	cmd := &cobra.Command{
		Use:   "checkPassphrase",
		Short: "Check session passphrase",
		Long:  `Check session passphrase`,
		Example: `
		`,
		RunE: ccmd.checkSession,
	}

	cmd.Flags().BoolVar(&ccmd.OutputJSON, "json", false, "Output passphrase in json")
	cmd.Flags().BoolVar(&ccmd.OutputEnvVariables, "env", false, "Output settings as shell text that can be imported using eval")
	cmd.SilenceUsage = true

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *checkSessionPassphraseCmd) savePassword(pass string) error {
	cliConfig.SetPassword(pass)
	return cliConfig.WritePersistentConfig()
}

func (n *checkSessionPassphraseCmd) checkSession(cmd *cobra.Command, args []string) error {
	// only create the key if the encryption is enabled
	encryptionEnabled := cliConfig.IsEncryptionEnabled()
	if encryptionEnabled {
		if err := cliConfig.ReadKeyFile(); err != nil {
			return err
		}

		// check if encryption is used on the current session
		passphrase, err := cliConfig.CheckEncryption()
		if err != nil {
			return err
		}
		if passphrase == "" || passphrase == "null" {
			return fmt.Errorf("passphrase can not be empty")
		}
		cliConfig.Passphrase = passphrase
	}

	if n.OutputJSON {
		cliConfig.Logger = Logger
		output := cliConfig.GetEnvironmentVariables(client, true)
		b, err := json.Marshal(output)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", b)
	} else if n.OutputEnvVariables {
		showEnvironmentVariables(nil, false)
	} else {
		n.showEnvironmentVariableUsage()
	}

	if encryptionEnabled {
		green := promptui.Styler(promptui.FGGreen)
		n.cmd.ErrOrStderr().Write([]byte(green("Passphrase OK\n")))
		Logger.Info("Passphrase accepted")
	}

	return nil
}

func (n *checkSessionPassphraseCmd) showEnvironmentVariableUsage() {
	envVar := "C8Y_PASSPHRASE"
	message := `
Powershell:

	$env:%[1]s = '%s'

bash/zsh:

	export %[1]s='%s'

`
	if os.Getenv(envVar) != cliConfig.Passphrase {
		fmt.Printf(
			message,
			envVar,
			strings.ReplaceAll(cliConfig.Passphrase, "'", "`'"),
			strings.ReplaceAll(cliConfig.Passphrase, "'", `\'`),
		)
	}
}
