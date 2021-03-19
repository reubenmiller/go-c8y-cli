package serviceuser

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/serviceuser/create"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/serviceuser/get"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdServiceUser struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdServiceUser {
	ccmd := &SubCmdServiceUser{}

	cmd := &cobra.Command{
		Use:   "serviceusers",
		Short: "Cumulocity serviceusers",
		Long:  `REST endpoint to interact with Cumulocity serviceusers`,
	}

	// Subcommands
	cmd.AddCommand(cmdCreate.NewCmdCreate(f).GetCommand())
	cmd.AddCommand(cmdGet.NewCmdGet(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
