// Package binaries wires the `c8y binaries` command and its subcommands. All
// subcommands are hand-written v2 c8ystream commands (backed by the go-c8y v2
// Binaries service), so this group command is itself hand-written rather than
// generated.
package binaries

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/binaries/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/binaries/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/binaries/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/binaries/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/binaries/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdBinaries struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdBinaries {
	ccmd := &SubCmdBinaries{}

	cmd := &cobra.Command{
		Use:   "binaries",
		Short: "Cumulocity binaries",
		Long:  `Manage binaries stored in the Cumulocity inventory`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
