package alarms

import (
	cmdCount "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alarms/count"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alarms/create"
	cmdDeleteCollection "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alarms/deletecollection"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alarms/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alarms/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alarms/update"
	cmdUpdateCollection "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alarms/updatecollection"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAlarms struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdAlarms {
	ccmd := &SubCmdAlarms{}

	cmd := &cobra.Command{
		Use:   "alarms",
		Short: "Cumulocity alarms",
		Long:  `REST endpoint to interact with Cumulocity alarms`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCount.NewCountCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdateCollection.NewUpdateCollectionCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteCollection.NewDeleteCollectionCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
