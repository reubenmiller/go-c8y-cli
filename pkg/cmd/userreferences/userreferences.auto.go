package userreferences

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	cmdAddUserToGroup "github.com/reubenmiller/go-c8y-cli/pkg/cmd/userreferences/addusertogroup"
	cmdDeleteUserFromGroup "github.com/reubenmiller/go-c8y-cli/pkg/cmd/userreferences/deleteuserfromgroup"
	cmdListGroupMembership "github.com/reubenmiller/go-c8y-cli/pkg/cmd/userreferences/listgroupmembership"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdUserreferences struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdUserreferences {
	ccmd := &SubCmdUserreferences{}

	cmd := &cobra.Command{
		Use:   "userreferences",
		Short: "Cumulocity user references",
		Long:  `REST endpoint to interact with Cumulocity user references`,
	}

	// Subcommands
	cmd.AddCommand(cmdAddUserToGroup.NewAddUserToGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdDeleteUserFromGroup.NewDeleteUserFromGroupCmd(f).GetCommand())
	cmd.AddCommand(cmdListGroupMembership.NewListGroupMembershipCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
