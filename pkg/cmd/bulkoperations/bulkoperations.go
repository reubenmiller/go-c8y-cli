// Package bulkoperations wires the `c8y bulkoperations` command group: the
// bulk-operation CRUD subcommands plus listOperations (the operations spawned
// by a bulk operation). All subcommands are hand-written v2 c8ystream commands
// backed by the go-c8y v2 BulkOperations service.
package bulkoperations

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdBulkoperations struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdBulkoperations {
	ccmd := &SubCmdBulkoperations{}

	cmd := &cobra.Command{
		Use:   "bulkoperations",
		Short: "Cumulocity bulk operations",
		Long:  `REST endpoint to interact with Cumulocity bulk operations`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewListOperationsCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
