package operations

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/operations/create"
	cmdDeleteCollection "github.com/reubenmiller/go-c8y-cli/pkg/cmd/operations/deletecollection"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/operations/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/operations/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/operations/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdOperations struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdOperations {
	ccmd := &SubCmdOperations{}

	cmd := &cobra.Command{
		Use:   "operations",
		Short: "Cumulocity operations",
		Long:  `REST endpoint to interact with Cumulocity operations`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteCollection.NewDeleteCollectionCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
