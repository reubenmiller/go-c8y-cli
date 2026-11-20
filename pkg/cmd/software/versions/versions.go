// Package versions wires the `c8y software versions` command and its
// subcommands. The subcommands are hand-written v2 c8ystream commands backed by
// the go-c8y v2 Repository.Software.Versions service (install/uninstall create
// device operations via Operations.CreateRaw). The create command is attached in
// pkg/cmd/root/root.go.
package versions

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdVersions struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdVersions {
	ccmd := &SubCmdVersions{}

	cmd := &cobra.Command{
		Use:   "versions",
		Short: "Cumulocity software version management",
		Long:  `Software version management to create/list/delete versions`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewInstallCmd(f).GetCommand())
	cmd.AddCommand(NewUninstallCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
