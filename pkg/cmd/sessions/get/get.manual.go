package get

import (
	"encoding/json"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type CmdGetSession struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdGetSession(f *cmdutil.Factory) *CmdGetSession {
	ccmd := &CmdGetSession{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get session",
		Long:  `Get session infomration and settings`,
		Example: heredoc.Doc(`
			$ c8y sessions get
			Get the details about the current session

			$ c8y sessions get --session mycustomsession
			Get the details about the current session which is specified via the --session argument

			$ c8y sessions get --select host,tenant
			Show the host and tenant name of the current session
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdGetSession) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	cfg.Persistent.Set("path", cfg.GetSessionFile())
	sessionFileContents := cfg.Persistent.AllSettings()

	b, err := json.Marshal(sessionFileContents)
	if err != nil {
		return err
	}
	return n.factory.WriteJSONToConsole(cfg, cmd, "", b)
}
