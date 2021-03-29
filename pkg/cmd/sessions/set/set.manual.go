package login

import (
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ylogin"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/factory"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/sessions/selectsession"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/shell"
	"github.com/reubenmiller/go-c8y-cli/pkg/utilities"
	"github.com/spf13/cobra"
)

type CmdSet struct {
	TFACode       string
	LoginErr      error
	LoginOK       bool
	Shell         string
	ClearToken    bool
	sessionFilter string

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdSet(f *cmdutil.Factory) *CmdSet {
	ccmd := &CmdSet{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set Cumulocity session",
		Long:  `Set a session, login and test the session and get either OAuth2 token, or using two factor authentication`,
		Example: heredoc.Doc(`
			$ eval $( c8y session set )
			Set a session interactively

			$ eval $( c8y sessions set --sessionFilter "company dev" )
			Set a session but only include session matching company AND dev

			$ eval $( c8y sessions set --session myfile.json --tfaCode 123456 )
			Set a session using a given file (non-interactively)
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	defaultShell := ""
	if !f.IOStreams.IsStdoutTTY() {
		defaultShell = "auto"
	}

	cmd.Flags().StringVar(&ccmd.sessionFilter, "sessionFilter", "", "Filter to be applied to the list of sessions even before the values can be selected")
	cmd.Flags().StringVar(&ccmd.TFACode, "tfaCode", "", "Two Factor Authentication code")
	cmd.Flags().StringVar(&ccmd.Shell, "shell", defaultShell, "Shell type to return the environment variables")
	cmd.Flags().BoolVar(&ccmd.ClearToken, "clear", false, "Clear any existing tokens")

	_ = cmd.MarkFlagRequired("shell")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("shell", "auto", "bash", "zsh", "fish", "powershell"),
	)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdSet) onSave() {
	cfg, _ := n.factory.Config()
	client, _ := n.factory.Client()
	log, _ := n.factory.Logger()
	log.Debug("Saving session file")

	if err := cfg.SaveClientConfig(client); err != nil {
		log.Errorf("Saving file error. %s", err)
	}
}

func (n *CmdSet) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}

	// cmd.Flags().GetString("session")
	sessionFile := cfg.GetSessionFile()
	// sessionFile := n.sessionFile
	if sessionFile == "" {
		sessionFile, err = selectsession.SelectSession(cfg, log, strings.Join(append(args, n.sessionFilter), " "))

		if err != nil {
			return err
		}
	}
	cfg.Logger.Debugf("selected session file: %s", sessionFile)
	if sessionFile != "" {
		// os.Clearenv()
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, "C8Y") && !strings.HasPrefix(env, "C8Y_PASSPHRASE") && !strings.HasPrefix(env, "C8Y_SESSION_HOME") {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					os.Unsetenv(parts[0])
				}
			}
		}
		// os.Setenv("C8Y_SESSION", sessionFile)
		cfg.SetSessionFile(sessionFile)
		_, _ = cfg.ReadConfigFiles(nil)
		// if err != nil {
		// 	return err
		// }
	}

	n.factory.Config = func() (*config.Config, error) {
		return cfg, nil
	}
	client, err := factory.CreateCumulocityClient(n.factory, "", "", "")()

	if err != nil {
		return err
	}

	if n.ClearToken {
		client.SetToken("")
	}

	if err := utilities.CheckEncryption(n.SubCommand.GetCommand().ErrOrStderr(), cfg, client); err != nil {
		return err
	}

	// If the password is not encrypted, then save it (which will apply the encryption)
	if !cfg.IsPasswordEncrypted() {
		if cfg.EncryptionEnabled() {
			log.Infof("Password is unencrypted. enforcing encryption")
			n.onSave()
		}
	}

	handler := c8ylogin.NewLoginHandler(client, cmd.ErrOrStderr(), n.onSave)

	handler.TFACode = n.TFACode
	handler.SetLogger(log)
	err = handler.Run()

	if err != nil {
		return err
	}

	// TODO: Persist password and/or tokens if enabled
	if handler.C8Yclient.TenantName != "" && cfg.GetTenant() != handler.C8Yclient.TenantName {
		log.Infof("Saving tenant name")
		n.onSave()
	}

	c8ysession.PrintSessionInfo(n.SubCommand.GetCommand().ErrOrStderr(), client, c8ysession.CumulocitySession{
		Path:     cfg.GetSessionFile(),
		Host:     handler.C8Yclient.BaseURL.Host,
		Tenant:   cfg.GetTenant(),
		Username: handler.C8Yclient.Username,
	})

	if n.Shell != "" {
		if strings.EqualFold(n.Shell, "auto") {
			n.Shell = shell.DetectShell("bash")
		}
		shell := utilities.ShellBash
		utilities.ShowClientEnvironmentVariables(cfg, handler.C8Yclient, shell.FromString(n.Shell))
	}

	return nil
}
