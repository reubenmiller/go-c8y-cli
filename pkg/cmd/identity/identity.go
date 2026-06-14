// Package identity wires the `c8y identity` command and its subcommands. Every
// subcommand (list/get/create/delete) is a hand-written v2 c8ystream command, so
// this group command is itself hand-written rather than generated from the API
// spec.
package identity

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/identity/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/identity/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/identity/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/identity/list"
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
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
