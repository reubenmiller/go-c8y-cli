// Package retentionrules wires the `c8y retentionrules` command and its
// subcommands. Every subcommand (list/get/create/update/delete) is a
// hand-written v2 c8ystream command, so this group command is itself
// hand-written rather than generated from the API spec.
package retentionrules

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/retentionrules/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/retentionrules/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/retentionrules/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/retentionrules/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/retentionrules/update"
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
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
