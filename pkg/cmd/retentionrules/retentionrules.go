// Package retentionrules wires the `c8y retentionrules` command and its
// subcommands. Every subcommand (list/get/create/update/delete) is a
// hand-written v2 c8ystream command, so this group command is itself
// hand-written rather than generated from the API spec.
package retentionrules

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// SubCmdRetentionrules is the `retentionrules` group command.
type SubCmdRetentionrules struct {
	*subcommand.SubCommand
}

// NewSubCommand builds the `retentionrules` group command and attaches its
// subcommands.
func NewSubCommand(f *cmdutil.Factory) *SubCmdRetentionrules {
	ccmd := &SubCmdRetentionrules{}

	cmd := &cobra.Command{
		Use:   "retentionrules",
		Short: "Cumulocity retentionRules",
		Long:  `REST endpoint to interact with Cumulocity retentionRules`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
