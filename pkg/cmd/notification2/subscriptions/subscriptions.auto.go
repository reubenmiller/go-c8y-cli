package subscriptions

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/notification2/subscriptions/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/notification2/subscriptions/delete"
	cmdDeleteBySource "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/notification2/subscriptions/deletebysource"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/notification2/subscriptions/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/notification2/subscriptions/list"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdSubscriptions struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdSubscriptions {
	ccmd := &SubCmdSubscriptions{}

	cmd := &cobra.Command{
		Use:   "subscriptions",
		Short: "Cumulocity notification2 subscriptions",
		Long:  `Methods to create, retrieve and delete notification subscriptions`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteBySource.NewDeleteBySourceCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
