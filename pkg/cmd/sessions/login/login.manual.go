package login

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/kballard/go-shellquote"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ylogin"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/prompt"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/shell"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CmdLogin struct {
	// Sources
	File     string
	Exec     string
	Stdin    bool
	Env      bool
	Prompt   bool
	Format   string
	Provider string
	Secrets  []string

	// Login options
	LoginType  string
	TFACode    string
	ClearToken bool

	Mode string
	Host string

	// SSO options
	SSOScopes   string
	SSOAudience string

	// Output options
	Shell        string
	OutputFormat string
	NoBanner     bool

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdLogin(f *cmdutil.Factory) *CmdLogin {
	ccmd := &CmdLogin{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "login to Cumulocity and return environment variables (including a token)",
		Long:  `Set a session, login and test the session and get either OAuth2 token, or using two factor authentication`,
		Example: heredoc.Doc(`
			$ eval "$( c8y sessions login --from-file .env )"
			Set a session from a dotenv file

			$ eval "$( c8y sessions login --from-env )"
			Set a session from environment variables (e.g. in Github)

			$ eval "$( c8y-session-bitwarden | c8y sessions login --from-stdin --format json )"
			Set a session from an external command, accepting the selected session via stdin

			$ eval "$( c8y sessions login --from-cmd "c8y sessions set --output json" )"
			Set a session using the in-built "c8y sessions set"

			$ eval "$( c8y sessions login --from-cmd "c8y-session-bitwarden list --folder c8y" --secrets BW_SESSION --format json )"
			Set a session from an external command, where the external commands returns the selected session in json format on stdout

			$ eval "$( c8y sessions login --from-prompt --host example.cumulocity.com)"
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.Provider, "provider", "", "Session provider which returns the session to use")
	cmd.Flags().StringVar(&ccmd.File, "from-file", "", "Read session from a file")
	cmd.Flags().StringVar(&ccmd.Exec, "from-cmd", "", "External command to execute to get the log in details")
	cmd.Flags().BoolVar(&ccmd.Env, "from-env", false, "Read from environment variables")
	cmd.Flags().BoolVar(&ccmd.Stdin, "from-stdin", false, "Read from standard input")
	cmd.Flags().BoolVar(&ccmd.Prompt, "from-prompt", false, "Read from user prompted input")
	cmd.Flags().BoolVar(&ccmd.NoBanner, "no-banner", false, "Don't show the session banner")
	cmd.Flags().StringVar(&ccmd.TFACode, "tfaCode", "", "Two Factor Authentication code")
	cmd.Flags().BoolVar(&ccmd.ClearToken, "clear", false, "Clear any existing tokens")
	cmd.Flags().StringVar(&ccmd.Format, "format", "", "External command format, e.g. json, yaml, toml")
	cmd.Flags().StringVar(&ccmd.OutputFormat, "output-format", "", "Output format")
	cmd.Flags().StringVar(&ccmd.Shell, "shell", "", "Shell type to return the environment variables")
	cmd.Flags().StringVar(&ccmd.LoginType, "loginType", "", "Login type preference, e.g. OAUTH2_INTERNAL, OAUTH2 (device flow) or BASIC. When set to BASIC, any existing token will be cleared")
	cmd.Flags().StringVar(&ccmd.Mode, "mode", "", "Session mode which controls which commands are allowed, e.g. dev, qual or prod")
	cmd.Flags().StringSliceVar(&ccmd.Secrets, "secrets", []string{}, "List of secrets to include as env variables when running an external command. Only valid with from-cmd")
	cmd.Flags().StringVar(&ccmd.Host, "host", "", "Cumulocity host. Only used with the 'interactive' provider")
	cmd.Flags().StringVar(&ccmd.SSOScopes, "sso-scopes", "", "SSO Scopes")
	cmd.Flags().StringVar(&ccmd.SSOAudience, "sso-audience", "", "SSO Audience")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("shell", shell.SupportedShells(shell.ShellAuto)...),
		completion.WithValidateSet("output-format", "json", "dotenv"),
		completion.WithValidateSet(
			"provider",
			config.ProviderTypeFile,
			config.ProviderTypeStdin,
			config.ProviderTypeEnv,
			config.ProviderTypeExternal,
			config.ProviderTypeAuto,
			config.ProviderTypeInteractive,
		),
		completion.WithValidateSet("format", "json", "yaml", "toml", "dotenv"),
		completion.WithValidateSet("loginType", c8y.LoginTypeOAuth2Internal, c8y.LoginTypeBasic, c8y.LoginTypeNone, c8y.LoginTypeOAuth2),
		completion.WithValidateSet(
			"mode",
			config.GetSessionModeCompletionHelp()...,
		),
	)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	cmd.MarkFlagsMutuallyExclusive("from-file", "from-cmd", "from-stdin", "from-env", "from-prompt")
	cmd.MarkFlagsMutuallyExclusive("output-format", "shell")

	return ccmd
}

func (n *CmdLogin) FromStdin(format string, args []string) (*c8ysession.CumulocitySession, error) {
	session, err := n.FromReader(bufio.NewReader(os.Stdin), format)

	if session.SessionUri == "" {
		session.SessionUri = "stdin://host"
	}
	return session, err
}

func (n *CmdLogin) FromEnv() (*c8ysession.CumulocitySession, error) {
	session := &c8ysession.CumulocitySession{
		Tenant:   os.Getenv("C8Y_TENANT"),
		Password: os.Getenv("C8Y_PASSWORD"),
		Token:    os.Getenv("C8Y_TOKEN"),
		Mode:     os.Getenv(config.EnvSessionMode),
	}

	// Choose the first non-empty value
	hostAliases := []string{
		"C8Y_HOST",
		"C8Y_URL",
		"C8Y_BASEURL",
	}
	for _, k := range hostAliases {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			session.SetHost(v)
			break
		}
	}

	// Username
	usernameAliases := []string{
		"C8Y_USERNAME",
		"C8Y_USER",
	}
	for _, k := range usernameAliases {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			session.Username = v
			break
		}
	}

	if session.SessionUri == "" {
		session.SessionUri = "env://host"
	}

	return session, nil
}

func (n *CmdLogin) FromInteractive(cmd *cobra.Command) (*c8ysession.CumulocitySession, error) {
	log, err := n.factory.Logger()
	if err != nil {
		return nil, err
	}

	prompter := prompt.NewPrompt(log)

	session := &c8ysession.CumulocitySession{
		Host: n.Host,
		Mode: n.Mode,
	}
	if v, err := cmd.Root().PersistentFlags().GetString("sessionUsername"); err == nil {
		session.Username = v
	}
	if v, err := cmd.Root().PersistentFlags().GetString("sessionPassword"); err == nil {
		session.Password = v
	}

	if session.Host == "" {
		v, err := prompter.Input("Enter host", "", true, false)
		if err != nil {
			return nil, err
		}
		session.Host = strings.TrimSpace(v)
	}

	if session.Mode == "" {
		mode, err := prompt.Select("Select mode", []string{
			"dev\tDevelopment mode (no restrictions)",
			"qual\tQA mode (delete disabled)",
			"prod\tProduction mode (read only)",
		}, "dev\tDevelopment mode (no restrictions)")
		if err != nil {
			return nil, err
		}
		session.Mode = mode
	}

	session.SessionUri = "interactive://host"
	return session, nil
}

func (n *CmdLogin) FromExternalProvider(args []string) (*c8ysession.CumulocitySession, error) {
	cfg, err := n.factory.Config()
	if err != nil {
		return nil, err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return nil, err
	}

	// add secrets when executing the environment in case
	// if the external command requires extra authentication
	env := os.Environ()
	for _, key := range n.Secrets {
		if v := os.Getenv(key); v == "" {
			secret, secretErr := cfg.PromptSecret(key)
			if secretErr != nil {
				if !errors.Is(secretErr, prompt.ErrNoPrompter) {
					cfg.Logger.Warnf("Could not get secret. key=%s, err=%s", key, secretErr)
				}
			} else {
				cfg.Logger.Debugf("Setting env variable secret for external provider. %s", key)
				env = append(env, fmt.Sprintf("%s=%s", key, secret))
			}
		}
	}

	providerCmd := strings.TrimSpace(n.Exec)
	if providerCmd == "" {
		return nil, fmt.Errorf("provider command is not set")
	}

	cmdArgs, err := shellquote.Split(strings.TrimSpace(providerCmd))
	if err != nil {
		return nil, err
	}
	if len(cmdArgs) == 0 {
		return nil, fmt.Errorf("provider command could not be parsed")
	}
	providerCommand := make([]string, 0, len(cmdArgs))
	providerCommand = append(providerCommand, cmdArgs...)

	if n.ClearToken {
		providerCommand = append(providerCommand, "--clear")
	}

	if n.LoginType != "" {
		providerCommand = append(providerCommand, fmt.Sprintf("--loginType=%s", n.LoginType))
	}

	if cfg.Verbose() {
		providerCommand = append(providerCommand, "-v")
	}

	providerCommand = append(providerCommand, args...)
	cmd := exec.Command(providerCommand[0], providerCommand[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	log.Infof("Executing session provider: %s", shellquote.Join(providerCommand...))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	if n.Format == "" {
		// Try to detect the format
		if jsonUtilities.IsJSONObject(output) {
			n.Format = "json"
			log.Infof("Detected input format: %s", n.Format)
		} else {
			n.Format = "dotenv"
			log.Infof("Guessing input format: %s", n.Format)
		}
	}

	log.Infof("Parsing session provider output: %s", output)
	return n.FromReader(bytes.NewReader(output), n.Format)
}

func (n *CmdLogin) FromViper(v *viper.Viper) (*c8ysession.CumulocitySession, error) {

	getValue := func(keys ...string) string {
		for _, k := range keys {
			if value := v.GetString(k); value != "" {
				return value
			}
			// use fallback value
			if value := v.GetString(config.EnvSettingsPrefix + "_" + k); value != "" {
				return value
			}
		}
		return ""
	}

	session := &c8ysession.CumulocitySession{
		SessionUri: getValue("sessionUri"),
		Path:       getValue("path"),
		Username:   getValue("username"),
		Password:   getValue("password"),
		Tenant:     getValue("tenant"),
		Token:      getValue("token"),
		TOTP:       getValue("totp"),
		Mode:       getValue("mode"),
		LoginType:  getValue("loginType"),
		Version:    getValue("version"),
	}

	// Get if value is defined before accessing it
	if v.GetString("authorized") != "" {
		session.SetAuthorized(v.GetBool("authorized"))
	}

	session.SetHost(getValue("host"))
	return session, nil
}

func (n *CmdLogin) FromFile(file string, format string) (*c8ysession.CumulocitySession, error) {
	v := viper.New()
	if format != "" {
		v.SetConfigType(format)
	}
	v.SetConfigFile(file)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	session, err := n.FromViper(v)
	if session.SessionUri == "" {
		session.SessionUri = "file://" + file
	}
	return session, err
}

func (n *CmdLogin) FromReader(r io.Reader, format string) (*c8ysession.CumulocitySession, error) {
	v := viper.New()
	if format != "" {
		v.SetConfigType(format)
	}
	err := v.ReadConfig(r)
	if err != nil {
		return nil, fmt.Errorf("invalid session format. expected_format=%s. error=%w", format, err)
	}

	return n.FromViper(v)
}

func oneHasChanged(cmd *cobra.Command, names ...string) bool {
	f := cmd.Flags()
	for _, name := range names {
		if f.Changed(name) {
			return true
		}
	}
	return false
}

func (n *CmdLogin) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}

	canChangeActiveSession := true
	// Warn users if they try to use this command directly
	if n.factory.IOStreams != nil {
		if n.factory.IOStreams.IsStdoutTTY() {
			canChangeActiveSession = false
			notice := heredoc.Docf(`
				You shouldn't run 'c8y session login' directly as it will have no effect on your current session.

				Instead, you will need to use the 'set-session' helper function, or if you can't use the helper function, then run:
		
				  # zsh/bash/sh
				  eval "$(c8y sessions login)"

				  # fish
				  c8y sessions login | source

				  # powershell
				  c8y sessions login | Out-String | Invoke-Expression
			`)
			fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s %s\n\n", strings.Repeat(n.factory.IOStreams.ColorScheme().WarningIcon(), 3), notice)
		}
	}

	if !oneHasChanged(cmd, "from-cmd", "from-file", "from-env", "from-stdin", "from-prompt") {

		// Set defaults from config if values from flags aren't provided
		if !cmd.Flags().Changed("provider") {
			n.Provider = cfg.SessionProvider()
			cfg.Logger.Debugf("Using session provider from configuration. type=%s", n.Provider)
		}

		n.Exec = cfg.SessionProviderCommand()
	}

	if !cmd.Flags().Changed("secrets") {
		n.Secrets = cfg.SessionProviderSecrets()
	}

	if n.Provider == "" {
		n.Provider = config.ProviderTypeAuto
	}

	// Clear any existing session file.
	// If the user wants to load from a file, then use the --from-file option
	cfg.ClearSessionFile()

	if strings.EqualFold(n.Provider, config.ProviderTypeAuto) {
		//
		// Try guessing a sensible default
		//
		if n.Prompt {
			n.Provider = config.ProviderTypeInteractive
		} else if n.File != "" {
			n.Provider = config.ProviderTypeFile
		} else if n.Stdin {
			n.Provider = config.ProviderTypeStdin
		} else if n.Env {
			n.Provider = config.ProviderTypeEnv
		} else if n.Exec != "" {
			n.Provider = config.ProviderTypeExternal
		} else if cfg.IsCIMode() {
			// CI environment and generally env variables are used here
			n.Provider = config.ProviderTypeEnv
		} else if n.factory.IOStreams.HasStdin() {
			n.Provider = config.ProviderTypeStdin
		}
	}

	var session *c8ysession.CumulocitySession

	switch strings.ToLower(n.Provider) {
	case config.ProviderTypeInteractive:
		session, err = n.FromInteractive(cmd)
	case config.ProviderTypeExternal:
		session, err = n.FromExternalProvider(args)
	case config.ProviderTypeEnv:
		session, err = n.FromEnv()
	case config.ProviderTypeStdin:
		if !n.factory.IOStreams.HasStdin() {
			err = fmt.Errorf("no stdin detected")
		} else {
			session, err = n.FromStdin(n.Format, args)
		}
	case config.ProviderTypeFile:
		session, err = n.FromFile(n.File, n.Format)
	default:
		return fmt.Errorf("unknown provider")
	}

	if err != nil {
		return err
	}

	// Fail early if the domain is not set
	if session.GetDomain() == "" {
		return cmderrors.NewUserError("invalid session. host is empty")
	}

	if sessionContents, err := json.Marshal(session); err == nil {
		cfg.Logger.Infof("Received session from external source:\n%s\n", sessionContents)
	}

	if session.SessionUri != "" {
		cfg.SetSessionFile(session.SessionUri)
	} else {
		cfg.SetSessionFile(session.Path)
	}

	client := c8y.NewClient(nil, session.Host, session.Tenant, session.Username, session.Password, true)

	if session.Token != "" {
		client.SetToken(session.Token)
	}

	if !session.IsAuthorized() {
		loginType := strings.ToUpper(cfg.GetLoginTypeWithDefault())
		if session.LoginType != "" {
			loginType = session.LoginType
		}
		if n.LoginType != "" {
			loginType = strings.ToUpper(n.LoginType)
		}
		log.Infof("User flag login type: %s", loginType)

		// Set default auth mode based on login type
		switch loginType {
		case c8y.LoginTypeOAuth2Internal, c8y.LoginTypeOAuth2:
			client.SetAuthorizationType(c8y.AuthTypeBearer)
		case c8y.LoginTypeBasic:
			client.SetAuthorizationType(c8y.AuthTypeBasic)
		case c8y.LoginTypeNone:
			client.SetAuthorizationType(c8y.AuthTypeNone)
		}

		// SSO providers are in control of the token, so it should just be used as is
		if !n.ClearToken && c8ysession.ShouldReuseToken(cfg, log, session.Token, loginType) {
			client.SetToken(session.Token)
		} else {
			client.ClearToken()
		}

		c8ysession.ClearProcessEnvironment()

		handler := c8ylogin.NewLoginHandler(n.factory.IOStreams, client, cmd.ErrOrStderr(), func() {})
		handler.LoginType = loginType
		handler.SSO.DiscoveryURL = cfg.SSODiscoveryUrl()

		if len(n.SSOScopes) > 0 {
			handler.SSO.Scopes = strings.Split(n.SSOScopes, " ")
		} else {
			handler.SSO.Scopes = cfg.SSOScopes()
		}

		handler.SSO.Audience = n.SSOAudience

		log.Infof("User preference for login type: %s", handler.LoginType)
		handler.TFACode = session.TOTP
		if n.TFACode == "" {
			if code, err := cfg.GetTOTP(time.Now()); err == nil {
				cfg.Logger.Infof("Setting totp code: %s", code)
				n.TFACode = code
			}
		}

		handler.SetLogger(log)
		err = handler.Run()
		if err != nil {
			return err
		}

		session.Username = handler.C8Yclient.Username
		session.Host = handler.C8Yclient.BaseURL.Host

		// Use the handler selected login type
		session.LoginType = handler.LoginType

		if client.Version != "" {
			session.Version = client.Version
		}

		if client.TenantName != "" {
			session.Tenant = client.TenantName
		}
		session.Token = client.Token
	}

	if n.Mode != "" {
		session.Mode = n.Mode
	}

	if session.Mode == "" || !config.IsSessionModeValid(session.Mode, cfg.IsCIMode()) {
		// prompt the user for a value
		modeOptions := config.GetSessionModeCompletionHelp()
		mode, selectErr := prompt.Select("Select session mode", modeOptions, modeOptions[0])
		if selectErr != nil {
			return selectErr
		}
		session.Mode = mode
	}

	// Write session details to stderr (for humans)
	cs := n.factory.IOStreams.ColorScheme()

	if canChangeActiveSession {
		fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s Session is now active\n", cs.SuccessIcon())
	} else {
		fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s Session is not active (see previous warning)\n", cs.WarningIcon())
	}
	if !n.NoBanner {
		c8ysession.PrintSessionInfo(n.SubCommand.GetCommand().ErrOrStderr(), client, cfg, *session)
	}

	outputFormat := n.OutputFormat
	if outputFormat == "" {
		if n.Shell == "" && !n.factory.IOStreams.IsStdoutTTY() {
			n.Shell = shell.ShellAuto
		}
		if strings.EqualFold(n.Shell, shell.ShellAuto) {
			n.Shell = shell.DetectShell(shell.ShellBash)
		}
		outputFormat = n.Shell
	}

	// Write session details to stdout (for machines)
	return c8ysession.WriteOutput(n.GetCommand().OutOrStdout(), client, cfg, session, outputFormat)
}
