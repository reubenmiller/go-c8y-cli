package login

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ylogin"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/factory"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/sessions/selectsession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/shell"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/utilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdSet struct {
	TFACode       string
	LoginErr      error
	LoginOK       bool
	LoginType     string
	Shell         string
	ClearToken    bool
	NoBanner      bool
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
			$ eval $( c8y sessions set )
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
		defaultShell = shell.ShellAuto
	}

	cmd.Flags().StringVar(&ccmd.sessionFilter, "sessionFilter", "", "Filter to be applied to the list of sessions even before the values can be selected")
	cmd.Flags().StringVar(&ccmd.TFACode, "tfaCode", "", "Two Factor Authentication code")
	cmd.Flags().StringVar(&ccmd.Shell, "shell", defaultShell, "Shell type to return the environment variables")
	cmd.Flags().StringVar(&ccmd.LoginType, "loginType", "", "Login type preference, e.g. OAUTH2_INTERNAL or BASIC. When set to BASIC, any existing token will be cleared")
	cmd.Flags().BoolVar(&ccmd.NoBanner, "no-banner", false, "Don't show the session banner")
	cmd.Flags().BoolVar(&ccmd.ClearToken, "clear", false, "Clear any existing tokens")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("shell", shell.SupportedShells(shell.ShellAuto)...),
		completion.WithValidateSet("loginType", c8y.AuthMethodOAuth2Internal, c8y.AuthMethodBasic, c8y.AuthMethodNone),
	)
	// Disable the encryption check, as the login handler will take care
	// of checking the encryption
	cmdutil.DisableEncryptionCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdSet) onSave(client *c8y.Client) {
	cfg, _ := n.factory.Config()
	log, _ := n.factory.Logger()

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

	// Ensure session values from previously read configuration is not transferred
	cfg.Set("settings.session.mode", nil)

	canChangeActiveSession := true
	// Warn users if they try to use this command directly
	if n.factory.IOStreams != nil {
		if n.factory.IOStreams.IsStdoutTTY() {
			canChangeActiveSession = false
			notice := heredoc.Docf(`
				You shouldn't run 'c8y session set' directly as it will have no effect on your current session.

				Instead, you will need to use the 'set-session' helper function, or if you can't use the helper function, then run:
		
				  # zsh/bash/sh
				  eval "$(c8y sessions set)"

				  # fish
				  c8y sessions set | source

				  # powershell
				  c8y sessions set | Out-String | Invoke-Expression
			`)
			fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s %s\n\n", strings.Repeat(n.factory.IOStreams.ColorScheme().WarningIcon(), 3), notice)
		}
	}

	sessionFile := ""
	if cmd.Root().PersistentFlags().Changed("session") {
		sessionFile = cfg.GetSessionFile()
	}

	if sessionFile == "" {
		sessionFile, err = selectsession.SelectSession(n.factory.IOStreams, cfg, log, strings.Join(append(args, n.sessionFilter), " "))

		if err != nil {
			return err
		}
	}
	cfg.Logger.Debugf("selected session file: %s", sessionFile)
	if sessionFile != "" {
		// Note: Ignore any environment variables as the session should take precedence because
		// the user is most likely switching session so does not want to inherit any environment variables
		// set from the last instance.
		// But this has a side effect that you can't control the profile handing via environment variables when using the interact session selection
		allowedEnvValues := []string{
			config.EnvSessionHide,
			// Preserve encryption settings
			config.EnvPassphrase,
			config.EnvPassphraseText,
			config.EnvSessionHome,
		}
		unsetEnvSettings := []string{
			cfg.GetEnvKey(config.SettingsLoginType),
			cfg.GetEnvKey(config.SettingsMode),
			cfg.GetEnvKey(config.EnvSessionMode),
		}
		env_prefix := strings.ToUpper(config.EnvSettingsPrefix)
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, env_prefix) {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					if cfg.HasEnvSettingsPrefix(parts[0]) && !slices.Contains(unsetEnvSettings, parts[0]) {
						continue
					}
					if !slices.Contains(allowedEnvValues, parts[0]) {
						os.Unsetenv(parts[0])
					}
				}
			}
		}

		// Clear existing token when using basic auth

		cfg.SetSessionFile(sessionFile)
		_, err = cfg.ReadConfigFiles(nil)
		if err != nil {
			return err
		}
	}

	n.factory.Config = func() (*config.Config, error) {
		return cfg, nil
	}
	client, err := factory.CreateCumulocityClient(n.factory, "", "", "", true)()

	if err != nil {
		return err
	}

	if n.LoginType == "" {
		n.LoginType = cfg.GetLoginTypeWithDefault()
	} else {
		if v, err := c8y.ParseAuthMethod(n.LoginType); err != nil {
			n.LoginType = c8y.AuthMethodOAuth2Internal
		} else {
			n.LoginType = v
		}
	}

	cfg.SetLoginType(n.LoginType)
	client.AuthorizationMethod = n.LoginType

	token := cfg.MustGetToken(true)
	if !n.ClearToken && c8ysession.ShouldReuseToken(cfg, log, token) {
		client.SetToken(token)
	} else {
		client.SetToken("")
	}

	if n.LoginType != c8y.AuthMethodNone {
		if err := utilities.CheckEncryption(n.factory.IOStreams, cfg, client); err != nil {
			return err
		}
	}

	// If the password is not encrypted, then save it (which will apply the encryption)
	if !cfg.IsPasswordEncrypted() {
		if cfg.EncryptionEnabled() {
			log.Infof("Password is unencrypted. enforcing encryption")
			n.onSave(nil)
		}
	}

	handler := c8ylogin.NewLoginHandler(client, cmd.ErrOrStderr(), func() {
		n.onSave(client)
	})
	handler.LoginType = n.LoginType
	log.Infof("User preference for login type: %s", handler.LoginType)

	if n.TFACode == "" {
		if code, err := cfg.GetTOTP(time.Now()); err == nil {
			cfg.Logger.Infof("Setting totp code: %s", code)
			n.TFACode = code
		}
	}
	handler.TFACode = n.TFACode
	handler.SetLogger(log)
	err = handler.Run()

	if err != nil {
		return err
	}

	mode := cfg.SessionMode(config.SessionModeUnset).String()
	// Respect session mode using the legacy format (e.g. set via individual create,update,delete values)
	if legacyMode, legacyValueExists := config.HasLegacySessionMode(cfg.Persistent); legacyValueExists {
		mode = legacyMode.String()
		cfg.Logger.Debugf("Detected legacy mode. mode=%s", mode)
	}

	if hasChanged(handler.C8Yclient, cfg) {
		log.Infof("Saving tenant name")
		n.onSave(handler.C8Yclient)
	}

	// include scheme so the user is aware if they are using http or https
	formatURL := strings.TrimSuffix(c8y.FormatBaseURL(handler.C8Yclient.BaseURL.String()), "/")
	session := &c8ysession.CumulocitySession{
		Path:       cfg.GetSessionFile(),
		SessionUri: "file://" + cfg.GetSessionFile(),
		Host:       formatURL,
		Password:   handler.C8Yclient.Password,
		Token:      handler.C8Yclient.Token,
		Tenant:     cfg.GetTenant(),
		Version:    cfg.GetCumulocityVersion(),
		Username:   handler.C8Yclient.Username,
		Mode:       mode,
	}

	outputFormat := cfg.GetOutputFormatWithDefault(cmd, config.OutputUnknown).String()

	// Write session details to stderr (for humans)
	if outputFormat != config.OutputJSON.String() {
		cs := n.factory.IOStreams.ColorScheme()
		if canChangeActiveSession {
			fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s Session is now active\n", cs.SuccessIcon())
		} else {
			fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s Session is not active (see previous warning)\n", cs.WarningIcon())
		}
		if !n.NoBanner {
			c8ysession.PrintSessionInfo(n.factory.IOStreams.ErrOut, client, cfg, *session)
		}
	}

	if outputFormat == config.OutputUnknown.String() {
		if n.Shell == "" && !n.factory.IOStreams.IsStdoutTTY() {
			n.Shell = shell.ShellAuto
		}
		if strings.EqualFold(n.Shell, shell.ShellAuto) {
			n.Shell = shell.DetectShell(shell.ShellBash)
		}
		outputFormat = n.Shell
	} else if outputFormat != config.OutputJSON.String() {
		// Don't output for any other format
		return nil
	}

	// Write session details to stdout (for machines)
	return c8ysession.WriteOutput(n.GetCommand().OutOrStdout(), client, cfg, session, outputFormat)
}

func hasChanged(client *c8y.Client, cfg *config.Config) bool {
	if client.TenantName != "" && client.TenantName != cfg.GetTenant() {
		return true
	}

	if client.Version != "" && client.Version != cfg.GetCumulocityVersion() {
		return true
	}

	if client.Token != "" && client.Token != cfg.MustGetToken(false) && cfg.StoreToken() {
		return true
	}

	if client.Password != "" && client.Password != cfg.MustGetPassword() && cfg.StorePassword() {
		return true
	}
	return false
}
