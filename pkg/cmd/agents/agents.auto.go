package cmd

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/agents/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/agents/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/agents/get"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/agents/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAgents struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdAgents {
	ccmd := &SubCmdAgents{}

	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Cumulocity agents",
		Long:  `REST endpoint to interact with Cumulocity agents`,
	}

	// Subcommands
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
