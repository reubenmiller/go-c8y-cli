package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type templateCmd struct {
	*subcommand.SubCommand
}

func NewTemplateRootCmd() *templateCmd {
	ccmd := &templateCmd{}

	cmd := &cobra.Command{
		Use:   "template",
		Short: "Template utilities",
		Long:  `Template utilities which can be used for testing out the templating language without sending any data to Cumulocity`,
	}

	// Subcommands
	cmd.AddCommand(newExecuteTemplateCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
