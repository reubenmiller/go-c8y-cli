package cmd

import (
	"fmt"

	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
)

type decryptTextCmd struct {
	passphrase string

	*baseCmd
}

func newDecryptTextCmd() *decryptTextCmd {
	ccmd := &decryptTextCmd{}

	cmd := &cobra.Command{
		Use:   "decryptText",
		Short: "Decrypt text",
		Long:  `Decrypt text based on the same encryption used to store sensitive data a cumulocity session`,
		Example: `
Example 1:
c8y session decryptText --text "{encrypted}asdfasdfasdfasdfasdf"

Encrypt the text "Hello World". You will be prompted for the passphrase to encrypt the data.

Example 2:
c8y session encryptText --text "Hello World" --passphrase "so4methIng-7hat-Matters"

Encrypt the text "Hello World", the text will be encrypted using the given passphrase (without being prompted)
		`,
		RunE: ccmd.decryptText,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("text", "", "Encrypted text. (required)")
	cmd.Flags().StringVar(&ccmd.passphrase, "passphrase", "", "Passphrase use for encoding your files")

	// Required flags
	_ = cmd.MarkFlagRequired("text")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *decryptTextCmd) decryptText(cmd *cobra.Command, args []string) error {

	if n.passphrase == "" {
		cmd.Printf("Enter password 🔒: [input is hidden] ")
		inputPassphrase, err := gopass.GetPasswd() // Silent
		if err != nil {
			return err
		}
		n.passphrase = string(inputPassphrase)
	}

	password := ""

	if v, err := cmd.Flags().GetString("text"); err == nil && v != "" {

		if cliConfig.SecureData.IsEncrypted(v) != 0 {
			password, err = cliConfig.SecureData.DecryptString(v, n.passphrase)
			if err != nil {
				return err
			}
		} else {
			Logger.Info("Text is already decrypted")
			password = v
		}
	}

	fmt.Printf("%s\n", password)
	return nil
}
