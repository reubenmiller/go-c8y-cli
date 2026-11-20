// Package currenttenant wires the `c8y currenttenant` command and its
// subcommands (get/listApplications/version). All are hand-written v2 c8ystream
// commands, so this group command is itself hand-written rather than generated.
package currenttenant

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdCurrenttenant struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdCurrenttenant {
	ccmd := &SubCmdCurrenttenant{}

	cmd := &cobra.Command{
		Use:   "currenttenant",
		Short: "Cumulocity current tenant",
		Long:  `Cumulocity current tenant commands`,
	}

	// Subcommands
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewListApplicationsCmd(f).GetCommand())
	cmd.AddCommand(NewVersionCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
