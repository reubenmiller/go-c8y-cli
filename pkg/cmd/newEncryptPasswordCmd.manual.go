package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type encryptPasswordCmd struct {
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

	// Required flags
	cmd.MarkFlagRequired("password")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *encryptPasswordCmd) encryptPassword(cmd *cobra.Command, args []string) error {

	session := &CumulocitySession{}
	if v, err := cmd.Flags().GetString("password"); err == nil && v != "" {
		session.SetPassword(v)
	}

	fmt.Printf("%0x", session.Password)

	return nil
	// return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "device", err))
}
