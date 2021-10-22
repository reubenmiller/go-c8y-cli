package assert

import (
	cmdText "github.com/reubenmiller/go-c8y-cli/pkg/cmd/assert/text"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAssert struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdAssert {
	ccmd := &SubCmdAssert{}

	cmd := &cobra.Command{
		Use:   "assert",
		Short: "Assertion utilities",
		Long:  `assertion utilities use for validating output`,
	}

	// Subcommands
	cmd.AddCommand(cmdText.NewCmdText(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
