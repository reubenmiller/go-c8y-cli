// Package firmware wires the `c8y firmware` command and its CRUD subcommands.
// The subcommands are hand-written v2 c8ystream commands backed by the go-c8y v2
// Repository.Firmware service, so this group command is itself hand-written
// rather than generated. The versions and patches subgroups (and their create
// commands) are attached in pkg/cmd/root/root.go.
package firmware

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdFirmware struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdFirmware {
	ccmd := &SubCmdFirmware{}

	cmd := &cobra.Command{
		Use:   "firmware",
		Short: "Cumulocity firmware management",
		Long:  `Firmware management to create/list/delete packages, versions and patches`,
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
