// Package events wires the `c8y events` command and its subcommands. All
// subcommands are hand-written v2 c8ystream commands (backed by the go-c8y v2
// Events service and its Binaries sub-service), so this group command is itself
// hand-written rather than generated.
package events

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdEvents struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdEvents {
	ccmd := &SubCmdEvents{}

	cmd := &cobra.Command{
		Use:   "events",
		Short: "Cumulocity events",
		Long:  `REST endpoint to interact with Cumulocity events`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCollectionCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewDownloadBinaryCmd(f).GetCommand())
	cmd.AddCommand(NewCreateBinaryCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateBinaryCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteBinaryCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
