package software

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/software/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/software/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/software/get"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/software/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
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
		Long:  `REST endpoint to interact with Cumulocity managed objects`,
	}

	// Subcommands
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
