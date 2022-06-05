package users

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/users/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/users/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/users/get"
	cmdGetInventoryRole "github.com/reubenmiller/go-c8y-cli/pkg/cmd/users/getinventoryrole"
	cmdGetUserByName "github.com/reubenmiller/go-c8y-cli/pkg/cmd/users/getuserbyname"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/users/list"
	cmdListInventoryRoles "github.com/reubenmiller/go-c8y-cli/pkg/cmd/users/listinventoryroles"
	cmdListUserMembership "github.com/reubenmiller/go-c8y-cli/pkg/cmd/users/listusermembership"
	cmdResetUserPassword "github.com/reubenmiller/go-c8y-cli/pkg/cmd/users/resetuserpassword"
	cmdRevokeTOTPSecret "github.com/reubenmiller/go-c8y-cli/pkg/cmd/users/revoketotpsecret"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/users/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdUsers struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdUsers {
	ccmd := &SubCmdUsers{}

	cmd := &cobra.Command{
		Use:   "users",
		Short: "Cumulocity users",
		Long:  `REST endpoint to interact with Cumulocity users`,
	}

	// Subcommands
	cmd.AddCommand(cmdListInventoryRoles.NewListInventoryRolesCmd(f).GetCommand())
	cmd.AddCommand(cmdGetInventoryRole.NewGetInventoryRoleCmd(f).GetCommand())
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdGetUserByName.NewGetUserByNameCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdResetUserPassword.NewResetUserPasswordCmd(f).GetCommand())
	cmd.AddCommand(cmdRevokeTOTPSecret.NewRevokeTOTPSecretCmd(f).GetCommand())
	cmd.AddCommand(cmdListUserMembership.NewListUserMembershipCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
