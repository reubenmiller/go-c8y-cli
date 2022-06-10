package events

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/events/create"
	cmdCreateBinary "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/events/createbinary"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/events/delete"
	cmdDeleteBinary "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/events/deletebinary"
	cmdDeleteCollection "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/events/deletecollection"
	cmdDownloadBinary "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/events/downloadbinary"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/events/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/events/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/events/update"
	cmdUpdateBinary "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/events/updatebinary"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdEvents struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdEvents {
	ccmd := &SubCmdEvents{}

	cmd := &cobra.Command{
		Use:   "events",
		Short: "Cumulocity events",
		Long:  `REST endpoint to interact with Cumulocity events`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteCollection.NewDeleteCollectionCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdDownloadBinary.NewDownloadBinaryCmd(f).GetCommand())
	cmd.AddCommand(cmdCreateBinary.NewCreateBinaryCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdateBinary.NewUpdateBinaryCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteBinary.NewDeleteBinaryCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
