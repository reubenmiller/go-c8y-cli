package sessions

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
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

	session, err := n.factory.ClientSession()
	if err != nil {
		return err
	}

	// Support looking up a session which is only controlled via env variables
	if !c8ysession.IsSessionFilePath(sessionPath) {
		host, username := session.Host, session.Username
		if session.Token != "" {
			// the token is the source of truth (it is the identity which is used for the requests)
			if claims, err := authentication.ParseToken(session.Token); err == nil {
				username = claims.User
				if host == "" || strings.Contains(host, claims.Issuer) {
					host = claims.Issuer
				}
			}
		}
		tenant := session.Tenant
		if tenant == "" {
			tenant = n.factory.GetTenant()
		}
		cfg.Persistent.Set("host", host)
		cfg.Persistent.Set("tenant", tenant)
		cfg.Persistent.Set("username", username)

		version := session.Version
		// a dry run does not send the request (the version stays unknown)
		if version == "" && !cfg.ShouldUseDryRun(cmd.CommandPath()) {
			client, err := n.factory.ClientV2()
			if err != nil {
				return err
			}
			res := client.Tenants.Current.GetVersion(context.Background())
			if res.Err != nil {
				return res.Err
			}
			version = res.Data.Value()
		}
		cfg.Persistent.Set("version", version)
	}

	sessionFileContents := cfg.Persistent.AllSettings()
	b, err := json.Marshal(sessionFileContents)
	if err != nil {
		return err
	}

	return n.factory.WriteOutputWithoutPropertyGuess(b, cmdutil.OutputContext{})
}
