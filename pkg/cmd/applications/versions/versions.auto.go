package versions

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/applications/versions/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/applications/versions/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/applications/versions/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/applications/versions/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/applications/versions/update"
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
		Short: "Cumulocity application versions",
		Long:  `API methods to retrieve, create, update and delete application versions`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
