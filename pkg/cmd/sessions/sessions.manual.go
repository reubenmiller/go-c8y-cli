package sessions

import (
	createCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/sessions/create"
	loginCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/sessions/login"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type sessionsCmd struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *sessionsCmd {
	ccmd := &sessionsCmd{}

	cmd := &cobra.Command{
		Use:     "sessions",
		Aliases: []string{"session"},
		Short:   "Cumulocity sessions",
		Long:    `Manage Cumulocity sessions`,
	}

	// Subcommands
	cmd.AddCommand(cmdutil.DisableAuthCheck(createCmd.NewCmdCreate(f).GetCommand()))
	cmd.AddCommand(NewCmdGetSession(f).GetCommand())
	cmd.AddCommand(cmdutil.DisableAuthCheck(NewCmdClearSession(f).GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(NewCmdDecryptText(f).GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(NewCmdEncryptText(f).GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(NewCmdList(f).GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(NewCmdSet(f).GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(NewCmdCloneSession(f).GetCommand()))
	cmd.AddCommand(cmdutil.DisableAuthCheck(loginCmd.NewCmdLogin(f).GetCommand()))

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
