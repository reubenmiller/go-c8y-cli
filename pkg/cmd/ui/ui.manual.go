package tenants

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdUI struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdUI {
	ccmd := &SubCmdUI{}

	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Cumulocity UI commands",
		Long:  `UI Commands`,
	}

	// Subcommands
	// cmd.AddCommand(cmdExtensions.NewSubCommand(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
