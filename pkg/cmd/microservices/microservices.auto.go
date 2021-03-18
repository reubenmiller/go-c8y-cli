package cmd

import (
	cmdCreateBinary "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/createbinary"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/delete"
	cmdDisable "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/disable"
	cmdEnable "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/enable"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/get"
	cmdGetBootstrapUser "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/getbootstrapuser"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdMicroservices struct {
	*subcommand.SubCommand
}

func NewSubCmdMicroservices(f *cmdutil.Factory) *SubCmdMicroservices {
	ccmd := &SubCmdMicroservices{}

	cmd := &cobra.Command{
		Use:   "microservices",
		Short: "Cumulocity microservices",
		Long:  `REST endpoint to interact with Cumulocity microservices`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdCreateBinary.NewCreateBinaryCmd(f).GetCommand())
	cmd.AddCommand(cmdGetBootstrapUser.NewGetBootstrapUserCmd(f).GetCommand())
	cmd.AddCommand(cmdEnable.NewEnableCmd(f).GetCommand())
	cmd.AddCommand(cmdDisable.NewDisableCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
