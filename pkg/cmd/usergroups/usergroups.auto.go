package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/usergroups/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/usergroups/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/usergroups/get"
	cmdGetByName "github.com/reubenmiller/go-c8y-cli/pkg/cmd/usergroups/getbyname"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/usergroups/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/usergroups/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdUsergroups struct {
	*subcommand.SubCommand
}

func NewSubCmdUsergroups(f *cmdutil.Factory) *SubCmdUsergroups {
	ccmd := &SubCmdUsergroups{}

	cmd := &cobra.Command{
		Use:   "usergroups",
		Short: "Cumulocity user groups",
		Long:  `REST endpoint to interact with Cumulocity user groups`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdGetByName.NewGetByNameCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
