package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	cmdAddRoleToGroup "github.com/reubenmiller/go-c8y-cli/pkg/cmd/userroles/addroletogroup"
	cmdAddRoleToUser "github.com/reubenmiller/go-c8y-cli/pkg/cmd/userroles/addroletouser"
	cmdDeleteRoleFromGroup "github.com/reubenmiller/go-c8y-cli/pkg/cmd/userroles/deleterolefromgroup"
	cmdDeleteRoleFromUser "github.com/reubenmiller/go-c8y-cli/pkg/cmd/userroles/deleterolefromuser"
	cmdGetRoleReferenceCollectionFromGroup "github.com/reubenmiller/go-c8y-cli/pkg/cmd/userroles/getrolereferencecollectionfromgroup"
	cmdGetRoleReferenceCollectionFromUser "github.com/reubenmiller/go-c8y-cli/pkg/cmd/userroles/getrolereferencecollectionfromuser"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/userroles/list"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdUserroles struct {
	*subcommand.SubCommand
}

func NewSubCmdUserroles(f *cmdutil.Factory) *SubCmdUserroles {
	ccmd := &SubCmdUserroles{}

	cmd := &cobra.Command{
		Use:   "userroles",
		Short: "Cumulocity user roles",
		Long:  `REST endpoint to interact with Cumulocity user roles`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdAddRoleToUser.NewAddRoleToUserCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteRoleFromUser.NewDeleteRoleFromUserCmd(f).GetCommand())
	cmd.AddCommand(cmdAddRoleToGroup.NewAddRoleToGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteRoleFromGroup.NewDeleteRoleFromGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdGetRoleReferenceCollectionFromUser.NewGetRoleReferenceCollectionFromUserCmd(f).GetCommand())
	cmd.AddCommand(cmdGetRoleReferenceCollectionFromGroup.NewGetRoleReferenceCollectionFromGroupCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
