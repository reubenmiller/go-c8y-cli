package cmd

import (
	"fmt"

	"github.com/howeyc/gopass"
	"github.com/reubenmiller/go-c8y-cli/pkg/encrypt"
	"github.com/spf13/cobra"
)

type encryptPasswordCmd struct {
	passphrase string

	*baseCmd
}

func newEncryptPasswordCmd() *encryptPasswordCmd {
	ccmd := &encryptPasswordCmd{}

	cmd := &cobra.Command{
		Use:   "encryptPassword",
		Short: "Create a new Cumulocity session credentials",
		Long:  `Create a new Cumulocity session credentials`,
		Example: `

		`,
		RunE: ccmd.encryptPassword,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("password", "", "Password. (required)")
	cmd.Flags().StringVar(&ccmd.passphrase, "passphrase", "", "Passphrase use for encoding your files")

	// Required flags
	cmd.MarkFlagRequired("password")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *encryptPasswordCmd) encryptPassword(cmd *cobra.Command, args []string) error {

	if n.passphrase == "" {
		cmd.Printf("Enter password 🔒: [input is hidden] ")
		inputPassphrase, err := gopass.GetPasswd() // Silent
		if err != nil {
			return err
		}
		n.passphrase = string(inputPassphrase)
	}

	session := &CumulocitySession{}
	if v, err := cmd.Flags().GetString("password"); err == nil && v != "" {
		session.SetPassword(encrypt.EncryptString(v, n.passphrase))
	}

	fmt.Printf("Password: %0x\n", session.Password)

	return nil
}
