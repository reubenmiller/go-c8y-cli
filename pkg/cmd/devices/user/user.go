// Package user wires the `c8y devices user` command and its subcommands. The
// get/update subcommands are hand-written v2 c8ystream commands backed by the
// go-c8y v2 ManagedObjects user endpoints, so this group command is itself
// hand-written rather than generated.
package user

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdUser struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdUser {
	ccmd := &SubCmdUser{}

	cmd := &cobra.Command{
		Use:   "user",
		Short: "Cumulocity device user management",
		Long:  `Managed the device user related to a device`,
	}

	// Subcommands
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
