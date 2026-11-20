// Package identity wires the `c8y identity` command and its subcommands. Every
// subcommand (list/get/create/delete/find) is a hand-written v2 c8ystream
// command, so this group command is itself hand-written rather than generated
// from the API spec.
package identity

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// SubCmdIdentity is the `identity` group command.
type SubCmdIdentity struct {
	*subcommand.SubCommand
}

// NewSubCommand builds the `identity` group command and attaches its
// subcommands.
func NewSubCommand(f *cmdutil.Factory) *SubCmdIdentity {
	ccmd := &SubCmdIdentity{}

	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Cumulocity external identity",
		Long:  `REST endpoint to interact with Cumulocity external identity objects`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewFindCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
