package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/howeyc/gopass"
	"github.com/reubenmiller/go-c8y-cli/pkg/encrypt"
	"github.com/spf13/cobra"
)

type decryptPasswordCmd struct {
	passphrase string

	*baseCmd
}

func newDecryptPasswordCmd() *decryptPasswordCmd {
	ccmd := &decryptPasswordCmd{}

	cmd := &cobra.Command{
		Use:   "decryptPassword",
		Short: "Create a new Cumulocity session credentials",
		Long:  `Create a new Cumulocity session credentials`,
		Example: `

		`,
		RunE: ccmd.decryptPassword,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("password", "", "Password. (required)")
	cmd.Flags().StringVar(&ccmd.passphrase, "passphrase", "", "Passphrase use for encoding your files")

	// Required flags
	cmd.MarkFlagRequired("password")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *decryptPasswordCmd) decryptPassword(cmd *cobra.Command, args []string) error {

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
		hexVal, err := hex.DecodeString(v)
		if err != nil {
			return err
		}
		pass, err := encrypt.Decrypt(hexVal, n.passphrase)
		if err != nil {
			return err
		}
		session.Password = string(pass)
	}

	fmt.Printf("%s\n", session.GetPassword())

	return nil
	// return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "device", err))
}
