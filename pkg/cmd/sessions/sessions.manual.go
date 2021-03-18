package sessions

import (
	clearCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/sessions/clear"
	createCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/sessions/create"
	decryptCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/sessions/decrypttext"
	encryptCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/sessions/encrypttext"
	getCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/sessions/get"
	listCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/sessions/list"
	loginCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/sessions/login"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type sessionsCmd struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *sessionsCmd {
	ccmd := &sessionsCmd{}

	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Cumulocity sessions",
		Long:  `Manage Cumulocity sessions`,
	}

	// Subcommands
	cmd.AddCommand(cmdutil.DisableAuthCheck(createCmd.NewCmdCreate(f).GetCommand()))
	cmd.AddCommand(getCmd.NewCmdGetSession(f).GetCommand())
	cmd.AddCommand(cmdutil.DisableAuthCheck(clearCmd.NewCmdClearSession(f).GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(decryptCmd.NewCmdDecryptText(f).GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(encryptCmd.NewCmdEncryptText(f).GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(listCmd.NewCmdList(f).GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(loginCmd.NewCmdLogin(f).GetCommand()))

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
