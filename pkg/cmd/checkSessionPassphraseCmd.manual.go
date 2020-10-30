package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/reubenmiller/go-c8y-cli/pkg/prompt"
	"github.com/spf13/cobra"
)

type checkSessionPassphraseCmd struct {
	prompter *prompt.Prompt
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
		RunE: ccmd.checkSession2,
	}

	cmd.SilenceUsage = true

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *checkSessionPassphraseCmd) savePassword(pass string) error {
	cliConfig.SetPassword(pass)
	return cliConfig.WritePersistentConfig()
}

func (n *checkSessionPassphraseCmd) checkSession2(cmd *cobra.Command, args []string) error {
	if err := cliConfig.ReadKeyFile(); err != nil {
		return err
	}

	// prompter := prompt.NewPrompt(Logger)
	Logger.Infof("SecretText: %s", cliConfig.SecretText)
	if err := cliConfig.CheckEncryption(); err != nil {
		return err
	}

	Logger.Info("Passphrase accepted")
	return nil
}

func (n *checkSessionPassphraseCmd) checkSession(cmd *cobra.Command, args []string) error {

	var pass string
	var c8ypass string
	var err error

	Logger.Debug("Checking passphrase")

	if cliConfig == nil {
		Logger.Errorf("CLI Config is nil")
		return nil
	}
	pass, err = cliConfig.GetPassword()

	if err == nil && pass == "" {
		// Prompt for new password
		c8ypass, err = n.prompter.Password("c8y")
		cliConfig.SetPassword(c8ypass)
		return n.savePassword(c8ypass)
	}

	if (err != nil || pass == "") && !cliConfig.IsCIMode() {
		pass, err = n.prompter.EncryptionPassphrase(cliConfig.SecretText)

		if err != nil {
			// Prompt for new password
			// Reset password
			confirm := promptui.Prompt{
				IsConfirm: true,
				Label:     "Do you want to set your saved Cumulocity Password again",
			}
			_, err = confirm.Run()

			if err != nil {
				Logger.Infof("User responded with no: %s", err)
				return err
			}

			c8ypass, err = n.prompter.Password("c8y")

			if err == nil {
				Logger.Infof("Saving data. passphrase=%s, password=%s", pass, c8ypass)
				cliConfig.Passphrase = pass
				n.savePassword(c8ypass)
			}
		}
	}

	if err != nil {
		Logger.Infof("Failed to decode password=%s", err)
		return err
	}

	Logger.Info("Successfully decoded password")
	n.showEnvironmentVariableUsage()
	return nil
}

func (n *checkSessionPassphraseCmd) showEnvironmentVariableUsage() {
	envVar := "C8Y_SESSION_PASSPHRASE"
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
