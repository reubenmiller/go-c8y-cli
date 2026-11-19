// Package alarms wires the `c8y alarms` command and its spec-derived
// subcommands. Every one is a hand-written v2 c8ystream command, so this group
// command is itself hand-written rather than generated. The CLI-only alarms
// subcommands (subscribe/assert) are attached on top of this command in
// pkg/cmd/root.
package alarms

import (
	cmdCount "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/alarms/count"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/alarms/create"
	cmdDeleteCollection "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/alarms/deletecollection"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/alarms/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/alarms/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/alarms/update"
	cmdUpdateCollection "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/alarms/updatecollection"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// SubCmdAlarms is the `alarms` group command.
type SubCmdAlarms struct {
	*subcommand.SubCommand
}

// NewSubCommand builds the `alarms` group command and attaches its spec-derived
// subcommands.
func NewSubCommand(f *cmdutil.Factory) *SubCmdAlarms {
	ccmd := &SubCmdAlarms{}

	cmd := &cobra.Command{
		Use:   "alarms",
		Short: "Cumulocity alarms",
		Long:  `REST endpoint to interact with Cumulocity alarms`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdateCollection.NewUpdateCollectionCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteCollection.NewDeleteCollectionCmd(f).GetCommand())
	cmd.AddCommand(cmdCount.NewCountCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
