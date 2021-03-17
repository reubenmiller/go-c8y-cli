package cmd

import (
	"fmt"

	"github.com/howeyc/gopass"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type encryptTextCmd struct {
	passphrase string
	raw        bool

	*subcommand.SubCommand
}

func newEncryptTextCmd() *encryptTextCmd {
	ccmd := &encryptTextCmd{}

	cmd := &cobra.Command{
		Use:   "encryptText",
		Short: "Encrypt text",
		Long:  `Encrypt text using the same encryption used for securely storing sensitive Cumulocity session information`,
		Example: `
Example 1: Encrypt the text "Hello World". You will be prompted for the passphrase to encrypt the data.

> c8y session encryptText --text "Hello World"
Enter password 🔒: [input is hidden] 
Password: {encrypted}ec5b837a03408ffb731307584eac40ac047989a002951e4b7139fa60189e504b6840bc027cece28b3f36717839d96af1c5dba8c850b9a9079846066ee1596cc8d26f4138f76ce3

Example 2: Encrypt the text "Hello World", the text will be encrypted using the given passphrase (without being prompted)

> c8y session encryptText --text "Hello World" --passphrase "so4methIng-7hat-Matters"
Password: {encrypted}ec5b837a03408ffb731307584eac40ac047989a002951e4b7139fa60189e504b6840bc027cece28b3f36717839d96af1c5dba8c850b9a9079846066ee1596cc8d26f4138f76ce3
		`,
		RunE: ccmd.encryptText,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("text", "", "Text to be encrypted. (required)")
	cmd.Flags().StringVar(&ccmd.passphrase, "passphrase", "", "Passphrase use for encrypting the text")
	cmd.Flags().BoolVar(&ccmd.raw, "raw", false, "Only return the encrypted text and nothing else")

	// Required flags
	_ = cmd.MarkFlagRequired("text")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *encryptTextCmd) encryptText(cmd *cobra.Command, args []string) error {

	if n.passphrase == "" {
		cmd.Printf("Enter password 🔒: [input is hidden] ")
		inputPassphrase, err := gopass.GetPasswd() // Silent
		if err != nil {
			return err
		}
		n.passphrase = string(inputPassphrase)
	}

	encryptedPassword := ""
	if v, err := cmd.Flags().GetString("text"); err == nil && v != "" {

		if cliConfig.SecureData.IsEncrypted(v) != 1 {
			data, err := cliConfig.SecureData.EncryptString(v, n.passphrase)

			if err != nil {
				return err
			}
			encryptedPassword = data
		} else {
			Logger.Info("Text is already encrypted")
			encryptedPassword = v
		}
	}

	if n.raw {
		fmt.Printf("%s\n", encryptedPassword)
	} else {
		fmt.Printf("Password: %s\n", encryptedPassword)
	}

	return nil
}
