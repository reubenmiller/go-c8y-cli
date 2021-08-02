package versions

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/versions/create"
	cmdCreatePatch "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/versions/createpatch"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/versions/delete"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/firmware/versions/list"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdVersions struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdVersions {
	ccmd := &SubCmdVersions{}

	cmd := &cobra.Command{
		Use:   "versions",
		Short: "Cumulocity firmware version management",
		Long:  `REST endpoint to interact with Cumulocity firmware versions`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdCreatePatch.NewCreatePatchCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
