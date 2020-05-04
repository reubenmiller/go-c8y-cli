package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type decryptPasswordCmd struct {
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

	// Required flags
	cmd.MarkFlagRequired("password")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *decryptPasswordCmd) decryptPassword(cmd *cobra.Command, args []string) error {

	session := &CumulocitySession{}
	if v, err := cmd.Flags().GetString("password"); err == nil && v != "" {
		session.Password = v
	}

	fmt.Printf("%s", session.GetPassword())

	return nil
	// return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "device", err))
}
