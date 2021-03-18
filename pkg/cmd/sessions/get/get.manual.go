package get

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/prompt"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

// CumulocitySessionDetails public details about the current session
type CumulocitySessionDetails struct {
	c8ysession.CumulocitySession

	Path string `json:"path"`
	Name string `json:"name"`
}

type CmdGetSession struct {
	OutputJSON bool
	prompter   *prompt.Prompt

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdGetSession(f *cmdutil.Factory) *CmdGetSession {
	ccmd := &CmdGetSession{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get session information",
		Long:  `Get session information`,
		Example: heredoc.Doc(`
Get the details about the current session
$ c8y sessions get

Get the details about the current session which is specified via the --session argument
$ c8y sessions get --session mycustomsession
		`),
		RunE: ccmd.RunE,
	}

	cmd.Flags().BoolVar(&ccmd.OutputJSON, "json", false, "Output passphrase in json")
	cmd.SilenceUsage = true

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdGetSession) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	n.prompter = prompt.NewPrompt(log)
	session := CumulocitySessionDetails{
		Path: cfg.GetSessionFilePath(),
		Name: cfg.GetName(),
		CumulocitySession: c8ysession.CumulocitySession{
			Host:            cfg.GetHost(),
			Tenant:          cfg.GetTenant(),
			Username:        cfg.GetUsername(),
			Description:     cfg.GetDescription(),
			UseTenantPrefix: cfg.Persistent.GetBool("useTenantPrefix"),

			Logger: log,
			Config: cfg,
		},
	}

	if session.CumulocitySession.Host == "" {
		return cmderrors.NewUserErrorWithExitCode(102, "no session loaded")
	}

	b, err := json.Marshal(session)
	if err != nil {
		return err
	}

	outputEnding := "\n"

	if n.OutputJSON {

		if cfg.CompactJSON() {
			fmt.Printf("%s%s", bytes.TrimSpace(b), outputEnding)
		} else {
			fmt.Printf("%s%s", pretty.Pretty(bytes.TrimSpace(b)), outputEnding)
		}
	} else {
		session.CumulocitySession.Path = session.Path
		c8ysession.PrintSessionInfo(n.SubCommand.GetCommand().ErrOrStderr(), client, session.CumulocitySession)
	}
	return nil
}
