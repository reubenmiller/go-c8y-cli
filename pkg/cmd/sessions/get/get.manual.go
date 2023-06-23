package get

import (
	"context"
	"encoding/json"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
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
		Long: heredoc.Doc(`
			Get session information and settings

			The session is either loaded by the session file or from environment variables.
		`),
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

	sessionPath := cfg.GetSessionFile()
	cfg.Persistent.Set("path", sessionPath)

	client, err := n.factory.Client()
	if err != nil {
		return err
	}

	// Support looking up a session which is only controlled via env variables
	if sessionPath == "" && client != nil {
		cfg.Persistent.Set("host", client.GetHostname())
		cfg.Persistent.Set("tenant", client.TenantName)
		cfg.Persistent.Set("username", client.GetUsername())

		version := client.Version
		if version == "" {
			serverVersion, err := client.TenantOptions.GetVersion(context.Background())
			if err != nil {
				return err
			}
			version = serverVersion
		}
		cfg.Persistent.Set("version", version)
	}

	sessionFileContents := cfg.Persistent.AllSettings()
	b, err := json.Marshal(sessionFileContents)
	if err != nil {
		return err
	}

	return n.factory.WriteJSONToConsole(cfg, cmd, "", b)
}
