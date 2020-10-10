package cmd

import (
	"github.com/spf13/cobra"
)

type templateCmd struct {
	*baseCmd
}

func newTemplateRootCmd() *templateCmd {
	ccmd := &templateCmd{}

	cmd := &cobra.Command{
		Use:   "template",
		Short: "Template utilities",
		Long:  `Template utilities which can be used for testing out the templating language without sending any data to Cumulocity`,
	}

	// Subcommands
	cmd.AddCommand(newExecuteTemplateCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
