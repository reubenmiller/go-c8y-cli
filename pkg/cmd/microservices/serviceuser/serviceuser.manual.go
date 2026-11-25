package serviceuser

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdServiceUser struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdServiceUser {
	ccmd := &SubCmdServiceUser{}

	cmd := &cobra.Command{
		Use:   "serviceusers",
		Short: "Cumulocity serviceusers",
		Long:  `REST endpoint to interact with Cumulocity serviceusers`,
	}

	// Subcommands
	cmd.AddCommand(NewCmdCreate(f).GetCommand())
	cmd.AddCommand(NewCmdGet(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
