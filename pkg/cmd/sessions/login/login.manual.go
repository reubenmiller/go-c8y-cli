package login

import (
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ylogin"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/utilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdLogin struct {
	TFACode              string
	LoginErr             error
	LoginOK              bool
	AsEnv                bool
	Shell                string
	ClearExistingCookies bool

	*subcommand.SubCommand

	factory *cmdutil.Factory
	Config  func() (*config.Config, error)
	Client  func() (*c8y.Client, error)
}

func NewCmdLogin(f *cmdutil.Factory) *CmdLogin {
	ccmd := &CmdLogin{
		factory: f,
		Config:  f.Config,
		Client:  f.Client,
	}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to cumulocity",
		Long:  `Login and test the Cumulocity session and get either OAuth2 token, or using two factor authentication`,
		Example: heredoc.Doc(`
c8y session login

Log into the current session
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.TFACode, "tfaCode", "", "Two Factor Authentication code")
	cmd.Flags().BoolVar(&ccmd.AsEnv, "env", false, "Return environment variables")
	cmd.Flags().StringVar(&ccmd.Shell, "shell", "bash", "Shell type")
	cmd.Flags().BoolVar(&ccmd.ClearExistingCookies, "clear", false, "Clear any existing cookies")

	_ = cmd.MarkFlagRequired("shell")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("shell", "bash", "zsh", "fish", "powershell"),
	)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdLogin) onSave() {
	cfg, _ := n.Config()
	client, _ := n.Client()
	log, _ := n.factory.Logger()
	log.Debug("Saving session file")

	if err := cfg.SaveClientConfig(client); err != nil {
		log.Errorf("Saving file error. %s", err)
	}
}

func (n *CmdLogin) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.Config()
	if err != nil {
		return err
	}
	client, err := n.Client()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	if n.ClearExistingCookies {
		client.SetCookies([]*http.Cookie{})
	}

	if err := utilities.CheckEncryption(n.SubCommand.GetCommand().ErrOrStderr(), cfg, client); err != nil {
		return err
	}

	// If the password is not encrypted, then save it (which will apply the encryption)
	if !cfg.IsPasswordEncrypted() {
		log.Infof("Password is unencrypted. enforcing encryption")
		n.onSave()
	}

	handler := c8ylogin.NewLoginHandler(client, cmd.ErrOrStderr(), n.onSave)

	handler.TFACode = n.TFACode
	handler.SetLogger(log)
	err = handler.Run()

	if err != nil {
		return err
	}

	if handler.C8Yclient.TenantName != "" && cfg.GetTenant() != handler.C8Yclient.TenantName {
		log.Infof("Saving tenant name")
		n.onSave()
	}

	c8ysession.PrintSessionInfo(n.SubCommand.GetCommand().ErrOrStderr(), client, c8ysession.CumulocitySession{
		Path:     cfg.GetSessionFilePath(),
		Host:     handler.C8Yclient.BaseURL.Host,
		Tenant:   cfg.GetTenant(),
		Username: handler.C8Yclient.Username,
	})

	if n.AsEnv {
		shell := utilities.ShellBash
		utilities.ShowClientEnvironmentVariables(cfg, handler.C8Yclient, shell.FromString(n.Shell))
	}

	return nil
}
