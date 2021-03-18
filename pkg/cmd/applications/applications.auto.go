package cmd

import (
	cmdCopy "github.com/reubenmiller/go-c8y-cli/pkg/cmd/applications/copy"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/applications/create"
	cmdCreateBinary "github.com/reubenmiller/go-c8y-cli/pkg/cmd/applications/createbinary"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/applications/delete"
	cmdDeleteApplicationBinary "github.com/reubenmiller/go-c8y-cli/pkg/cmd/applications/deleteapplicationbinary"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/applications/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/applications/list"
	cmdListApplicationBinaries "github.com/reubenmiller/go-c8y-cli/pkg/cmd/applications/listapplicationbinaries"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/applications/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdApplications struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdApplications {
	ccmd := &SubCmdApplications{}

	cmd := &cobra.Command{
		Use:   "applications",
		Short: "Cumulocity applications",
		Long:  `REST endpoint to interact with Cumulocity applications`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdCopy.NewCopyCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdCreateBinary.NewCreateBinaryCmd(f).GetCommand())
	cmd.AddCommand(cmdListApplicationBinaries.NewListApplicationBinariesCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteApplicationBinary.NewDeleteApplicationBinaryCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
