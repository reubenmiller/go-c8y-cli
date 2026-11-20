// Package features wires the `c8y features` command and its subcommands. All
// subcommands (the toggles here plus the tenants subgroup attached in
// root.go) are hand-written v2 c8ystream commands, so this group command is
// itself hand-written rather than generated.
package features

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdFeatures struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdFeatures {
	ccmd := &SubCmdFeatures{}

	cmd := &cobra.Command{
		Use:   "features",
		Short: "Cumulocity tenant features",
		Long:  `Managed feature toggles`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewEnableCmd(f).GetCommand())
	cmd.AddCommand(NewDisableCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
