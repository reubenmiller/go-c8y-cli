// Package operations wires the `c8y operations` command and its spec-derived
// subcommands (list/get/create/update/cancel/deleteCollection). Every one is a
// hand-written v2 c8ystream command, so this group command is itself
// hand-written rather than generated. The CLI-only operations subcommands
// (subscribe/wait/assert) are attached on top of this command in pkg/cmd/root.
package operations

import (
	cmdCancel "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/operations/cancel"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/operations/create"
	cmdDeleteCollection "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/operations/deletecollection"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/operations/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/operations/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/operations/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// SubCmdOperations is the `operations` group command.
type SubCmdOperations struct {
	*subcommand.SubCommand
}

// NewSubCommand builds the `operations` group command and attaches its
// spec-derived subcommands.
func NewSubCommand(f *cmdutil.Factory) *SubCmdOperations {
	ccmd := &SubCmdOperations{}

	cmd := &cobra.Command{
		Use:   "operations",
		Short: "Cumulocity operations",
		Long:  `REST endpoint to interact with Cumulocity operations`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdCancel.NewCancelCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteCollection.NewDeleteCollectionCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
