// Package additions wires the deprecated `c8y inventory additions` command and
// its subcommands (use `c8y inventory children ... --childType addition`). All
// subcommands are hand-written v2 c8ystream commands backed by the go-c8y v2
// ManagedObjects ChildAdditions service, so this group command is itself
// hand-written rather than generated.
package additions

import (
	cmdAssign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/additions/assign"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/additions/create"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/additions/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/additions/list"
	cmdUnassign "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/additions/unassign"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAdditions struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdAdditions {
	ccmd := &SubCmdAdditions{}

	cmd := &cobra.Command{
		Use:    "additions",
		Short:  "Cumulocity managed object additions",
		Long:   `Managed additions to managed objects`,
		Hidden: true,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdAssign.NewAssignCmd(f).GetCommand())
	cmd.AddCommand(cmdUnassign.NewUnassignCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
