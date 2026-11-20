// Package applications wires the `c8y applications` command and its subcommands.
// The CRUD + binary subcommands are hand-written v2 c8ystream commands (backed by
// the go-c8y v2 Applications service), so this group command is itself
// hand-written rather than generated. The versions subgroup and the manual
// createHostedApplication/open commands are attached in pkg/cmd/root/root.go.
package applications

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdApplications struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdApplications {
	ccmd := &SubCmdApplications{}

	cmd := &cobra.Command{
		Use:   "applications",
		Short: "Cumulocity applications",
		Long:  `REST endpoint to interact with Cumulocity applications`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewCopyCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewCreateBinaryCmd(f).GetCommand())
	cmd.AddCommand(NewListApplicationBinariesCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteApplicationBinaryCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
