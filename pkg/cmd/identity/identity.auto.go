package identity

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/identity/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/identity/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/identity/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/identity/list"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdIdentity struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdIdentity {
	ccmd := &SubCmdIdentity{}

	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Cumulocity identity",
		Long:  `REST endpoint to interact with Cumulocity identity objects`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
