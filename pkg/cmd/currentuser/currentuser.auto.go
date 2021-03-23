package currentuser

import (
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/currentuser/get"
	cmdLogout "github.com/reubenmiller/go-c8y-cli/pkg/cmd/currentuser/logout"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/currentuser/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdCurrentuser struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdCurrentuser {
	ccmd := &SubCmdCurrentuser{}

	cmd := &cobra.Command{
		Use:   "currentuser",
		Short: "Cumulocity current user",
		Long:  `REST endpoint to interact with the current Cumulocity user`,
	}

	// Subcommands
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdLogout.NewLogoutCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
