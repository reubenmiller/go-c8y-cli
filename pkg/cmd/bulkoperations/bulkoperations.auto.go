package bulkoperations

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/bulkoperations/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/bulkoperations/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/bulkoperations/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/bulkoperations/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/bulkoperations/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdBulkoperations struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdBulkoperations {
	ccmd := &SubCmdBulkoperations{}

	cmd := &cobra.Command{
		Use:   "bulkoperations",
		Short: "Cumulocity bulk operations",
		Long:  `REST endpoint to interact with Cumulocity bulk operations`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
