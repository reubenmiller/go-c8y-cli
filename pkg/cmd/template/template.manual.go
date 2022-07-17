package template

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	cmdExecute "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/template/execute"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdTemplate struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdTemplate {
	ccmd := &SubCmdTemplate{}

	cmd := &cobra.Command{
		Use:   "template",
		Short: "Template utilities",
		Long:  `Template utilities which can be used for testing out the templating language without sending any data to Cumulocity`,
	}

	// Subcommands
	cmd.AddCommand(cmdExecute.NewCmdExecute(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
