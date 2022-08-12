package software

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/software/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/software/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/software/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/software/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/software/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
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
		Long:  `Software management to create/list/delete packages and versions`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
