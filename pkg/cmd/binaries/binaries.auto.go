package binaries

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/binaries/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/binaries/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/binaries/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/binaries/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/binaries/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdBinaries struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdBinaries {
	ccmd := &SubCmdBinaries{}

	cmd := &cobra.Command{
		Use:   "binaries",
		Short: "Cumulocity binaries",
		Long:  `REST endpoint to interact with Cumulocity binaries`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
