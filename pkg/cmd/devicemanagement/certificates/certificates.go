// Package certificates wires the `c8y devicemanagement certificates` command and
// its subcommands. All subcommands are hand-written v2 c8ystream commands (backed
// by the go-c8y v2 TrustedCertificates service), so this group command is itself
// hand-written rather than generated.
package certificates

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdCertificates struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdCertificates {
	ccmd := &SubCmdCertificates{}

	cmd := &cobra.Command{
		Use:   "certificates",
		Short: "Device Certificate management",
		Long:  `Manage the trusted certificates which are used by devices.`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
