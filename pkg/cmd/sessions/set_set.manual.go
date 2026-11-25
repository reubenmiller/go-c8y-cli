package sessions

import (
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ylogin"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/sessions/selectsession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logintype"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/shell"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/utilities"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CmdSet struct {
	TFACode            string
	LoginErr           error
	LoginOK            bool
	LoginType          string
	Shell              string
	ClearToken         bool
	NoBanner           bool
	BrowserCallbackURL string
	sessionFilter      string

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

			$ eval $( c8y sessions set --session myfile.json --loginType DEVICE )
			Login using the OAuth2 device flow (polls for approval after visiting the URL shown in the terminal)

			$ eval $( c8y sessions set --session myfile.json --loginType BROWSER )
			Login using Authorization Code flow – opens the system browser and waits for the callback

			$ eval $( c8y sessions set --session myfile.json --loginType BROWSER --browserCallback http://127.0.0.1:8080/callback )
			Login via browser with a custom callback URL registered in the SSO provider

			$ eval $( c8y sessions set --session myfile.json --loginType CERTIFICATE )
			Login using a device certificate (mTLS); reads cert paths from the session file or C8Y_CERTIFICATE / C8Y_CERTIFICATE_KEY env vars
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
	cmd.Flags().StringVar(&ccmd.LoginType, "loginType", "", "Login type preference, e.g. OAUTH2_INTERNAL, DEVICE (device flow), BROWSER (Authorization Code via browser), BASIC or CERTIFICATE (mTLS). OAUTH2 is a legacy alias for DEVICE. When set to BASIC, any existing token will be cleared")
	cmd.Flags().BoolVar(&ccmd.NoBanner, "no-banner", false, "Don't show the session banner")
	cmd.Flags().BoolVar(&ccmd.ClearToken, "clear", false, "Clear any existing tokens")
	cmd.Flags().StringVar(&ccmd.BrowserCallbackURL, "browserCallback", "", "Custom redirect URI for the browser flow local callback server, e.g. http://127.0.0.1:8080/callback. Must match a URI registered in the SSO provider. Defaults to http://127.0.0.1:5001/callback")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("shell", shell.SupportedShells(shell.ShellAuto)...),
		completion.WithValidateSet("loginType", logintype.OAuth2Internal, logintype.Basic, logintype.None, logintype.OAuth2, logintype.Device, logintype.Browser, logintype.Certificate),
	)
	// Disable the encryption check, as the login handler will take care
	// of checking the encryption
	cmdutil.DisableEncryptionCheck(cmd)
	cmdutil.MarkLocalCommand(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdSet) onSave(client *config.ClientSession) {
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

	if n.LoginType == "" {
		n.LoginType = cfg.GetLoginTypeWithDefault()
	} else {
		n.LoginType = strings.ToUpper(n.LoginType)
	}
	if n.LoginType != "" && n.LoginType != logintype.Browser && n.LoginType != logintype.Certificate && n.LoginType != logintype.Device {
		// ParseLoginType only knows the types of the SDK; skip it for the CLI only values.
		if v, err := authentication.ParseLoginType(n.LoginType); err != nil {
			log.Warnf("Could not parse auth method: value=%s", err)
		} else {
			n.LoginType = v
		}
	}

	if n.LoginType != "" {
		cfg.SetLoginType(n.LoginType)
	}
	// Note: Always bind the environment variables of the session, even if the session was selected
	if err := cfg.BindAuthorization(); err != nil {
		log.Warnf("Failed to bind to authorization variables. %s", err)
	}
	clientOpts, err := n.factory.ClientOptionsV2(true)
	if err != nil {
		return err
	}
	baseURL, _ := url.Parse(clientOpts.BaseURL)
	client := config.NewClientSession(baseURL, clientOpts.Auth, "")
	if viper.IsSet("useTenantPrefix") {
		client.OmitTenant = !viper.GetBool("useTenantPrefix")
	}

	if n.LoginType != logintype.None {
		if err := utilities.CheckEncryption(n.factory.IOStreams, cfg, &client); err != nil {
			return err
		}
	}

	// Note: Decrypt after checking for encryption
	token := cfg.MustGetToken(true)
	if n.ClearToken {
		client.Token = ""
		cfg.ClearToken()
	} else if c8ysession.ShouldReuseToken(cfg, log, token, n.LoginType) {
		client.Token = token
		client.AuthType = authentication.AuthTypeBearer
	} else {
		client.Token = ""
		cfg.ClearToken()
	}

	// If the password is not encrypted, then save it (which will apply the encryption)
	if !cfg.IsPasswordEncrypted() {
		if cfg.EncryptionEnabled() {
			log.Infof("Password is unencrypted. enforcing encryption")
			n.onSave(nil)
		}
	}

	if n.TFACode == "" {
		if code, err := cfg.GetTOTP(time.Now()); err == nil {
			cfg.Logger.Infof("Setting totp code: %s", code)
			n.TFACode = code
		}
	}

	var handler *c8ylogin.LoginHandlerV2
	handler = c8ylogin.NewLoginHandlerV2(
		n.factory.IOStreams,
		apiv2.ClientOptions{
			BaseURL: clientOpts.BaseURL,
			Auth: authentication.AuthOptions{
				Tenant:   client.Tenant,
				Username: client.Username,
				Password: client.Password,
				Token:    client.Token,
				AuthType: []authentication.AuthType{client.AuthType},
			},
			InsecureSkipVerify: clientOpts.InsecureSkipVerify,
			Transport:          clientOpts.Transport,
		},
		cmd.ErrOrStderr(),
		func() {
			session := handler.Session()
			session.OmitTenant = client.OmitTenant
			n.onSave(&session)
		},
	)
	handler.LoginType = n.LoginType
	if n.LoginType != "" {
		handler.AllowedLoginTypes = strings.Split(n.LoginType, ",")
	}
	handler.CertificatePath = cfg.GetCertificate()
	handler.CertificateKeyPath = cfg.GetCertificateKey()
	handler.SSO.DiscoveryURL = cfg.SSODiscoveryUrl()
	handler.SSO.Scopes = cfg.SSOScopes()
	handler.BrowserCallbackURL = n.BrowserCallbackURL
	if handler.BrowserCallbackURL == "" {
		handler.BrowserCallbackURL = cfg.BrowserCallbackURL()
	}
	handler.TOTPCode = n.TFACode
	handler.SetLogger(log)
	if handler.LoginType == "" {
		log.Infof("User preference for login type: %s", "not-set")
	} else {
		log.Infof("User preference for login type: %s", handler.LoginType)
	}
	if err := handler.Run(); err != nil {
		return err
	}
	handlerLoginType := handler.LoginType
	client = handler.Session()
	if viper.IsSet("useTenantPrefix") {
		client.OmitTenant = !viper.GetBool("useTenantPrefix")
	}

	mode := cfg.SessionMode(config.SessionModeUnset).String()
	// Respect session mode using the legacy format (e.g. set via individual create,update,delete values)
	if legacyMode, legacyValueExists := config.HasLegacySessionMode(cfg.Persistent); legacyValueExists {
		mode = legacyMode.String()
		cfg.Logger.Debugf("Detected legacy mode. mode=%s", mode)
	}

	if hasChanged(&client, cfg) {
		log.Infof("Saving tenant name")
		n.onSave(&client)
	}

	// include scheme so the user is aware if they are using http or https
	formatURL := strings.TrimSuffix(apiv2.FormatBaseURL(client.URL), "/")
	session := &c8ysession.CumulocitySession{
		Path:       cfg.GetSessionFile(),
		SessionUri: "file://" + cfg.GetSessionFile(),
		Host:       formatURL,
		Password:   client.Password,
		Token:      client.Token,
		Tenant:     cfg.GetTenant(),
		Version:    cfg.GetCumulocityVersion(),
		Username:   client.Username,
		Mode:       mode,
	}
	// Don't trigger another login afterwards
	session.SetAuthorized(true)

	// Store the chosen login type for display to the user
	session.LoginType = handlerLoginType

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
			c8ysession.PrintSessionInfo(n.factory.IOStreams.ErrOut, &client, cfg, *session)
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
	cfg.Logger.Infof("c8y sessions set: isAuthorized: %v", session.IsAuthorized())
	return c8ysession.WriteOutput(n.GetCommand().OutOrStdout(), &client, cfg, session, outputFormat)
}

func hasChanged(client *config.ClientSession, cfg *config.Config) bool {
	if client.Tenant != "" && client.Tenant != cfg.GetTenant() {
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

	if client.Username != "" && client.Username != cfg.GetUsername() {
		return true
	}
	return false
}
