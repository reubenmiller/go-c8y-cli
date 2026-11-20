// Package software wires the `c8y software` command and its CRUD subcommands.
// The subcommands are hand-written v2 c8ystream commands backed by the go-c8y v2
// Repository.Software service, so this group command is itself hand-written
// rather than generated. The versions subgroup (and its create command) is
// attached in pkg/cmd/root/root.go.
package software

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdSoftware struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdSoftware {
	ccmd := &SubCmdSoftware{}

	cmd := &cobra.Command{
		Use:   "software",
		Short: "Cumulocity software management",
		Long:  `Software management to create/list/delete packages and versions`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
