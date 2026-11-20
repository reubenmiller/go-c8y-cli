// Package microservices wires the `c8y microservices` command and its
// subcommands. The CRUD + enable/disable/status/binary subcommands are
// hand-written v2 c8ystream commands (backed by the go-c8y v2 Microservices
// service), so this group command is itself hand-written rather than generated.
// The loglevels subgroup and the manual create/serviceuser commands are
// attached in pkg/cmd/root/root.go.
package microservices

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdMicroservices struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdMicroservices {
	ccmd := &SubCmdMicroservices{}

	cmd := &cobra.Command{
		Use:   "microservices",
		Short: "Cumulocity microservices",
		Long:  `REST endpoint to interact with Cumulocity microservices`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewCreateBinaryCmd(f).GetCommand())
	cmd.AddCommand(NewGetBootstrapUserCmd(f).GetCommand())
	cmd.AddCommand(NewEnableCmd(f).GetCommand())
	cmd.AddCommand(NewDisableCmd(f).GetCommand())
	cmd.AddCommand(NewGetStatusCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
