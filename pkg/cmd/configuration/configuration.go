// Package configuration wires the `c8y configuration` command and its
// subcommands. All subcommands are hand-written v2 c8ystream commands (backed by
// the go-c8y v2 Configuration service), so this group command is itself
// hand-written rather than generated.
package configuration

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdConfiguration struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdConfiguration {
	ccmd := &SubCmdConfiguration{}

	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Cumulocity configuration repository management",
		Long:  `Configuration management to create/list/delete configurations`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewSendCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
