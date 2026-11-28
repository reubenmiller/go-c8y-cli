package login

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
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
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logintype"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/prompt"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/shell"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
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

	// FromExtension is the session provider of an extension (<provider> or <extension>/<provider>)
	FromExtension string

	// Select is the sessionUri of the session which the extension session provider returns,
	// without user interaction (see the provider's select command)
	Select string

	// Login options
	LoginType          string
	TFACode            string
	ClearToken         bool
	BrowserCallbackURL string

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

			$ eval "$( c8y sessions login --from-extension bitwarden )"
			Set a session using the session provider "bitwarden" of an installed extension

			$ eval "$( c8y sessions login --from-extension bitwarden --select "bitwarden://3f0b9d4e-1c2a-4f5b-9e8d-7a6b5c4d3e2f" )"
			Set a session from the session provider "bitwarden" without selecting it interactively (the provider must declare list and select)

			$ eval "$( c8y sessions login --from-prompt --host example.cumulocity.com)"

			$ eval "$( c8y sessions login --from-file .env --loginType DEVICE )"
			Login using the OAuth2 device flow (polls for approval after visiting the URL shown in the terminal)

			$ eval "$( c8y sessions login --from-file .env --loginType BROWSER )"
			Login using Authorization Code flow – opens the system browser and waits for the callback

			$ eval "$( c8y sessions login --from-file .env --loginType BROWSER --browserCallback http://127.0.0.1:8080/callback )"
			Login via browser with a custom callback URL registered in the SSO provider

			$ eval "$( c8y sessions login --from-file .env --loginType CERTIFICATE )"
			Login using a device certificate (mTLS); reads cert paths from the session file or C8Y_CERTIFICATE / C8Y_CERTIFICATE_KEY env vars
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.Provider, "provider", "", "Session provider which returns the session to use")
	cmd.Flags().StringVar(&ccmd.File, "from-file", "", "Read session from a file")
	cmd.Flags().StringVar(&ccmd.Exec, "from-cmd", "", "External command to execute to get the log in details")
	cmd.Flags().StringVar(&ccmd.FromExtension, "from-extension", "", "Session provider of an installed extension, i.e. <provider> or <extension>/<provider>")
	cmd.Flags().StringVar(&ccmd.Select, "select", "", "sessionUri of the session to use from the extension session provider, without user interaction. Only valid with from-extension (or the extension provider type)")
	cmd.Flags().BoolVar(&ccmd.Env, "from-env", false, "Read from environment variables")
	cmd.Flags().BoolVar(&ccmd.Stdin, "from-stdin", false, "Read from standard input")
	cmd.Flags().BoolVar(&ccmd.Prompt, "from-prompt", false, "Read from user prompted input")
	cmd.Flags().BoolVar(&ccmd.NoBanner, "no-banner", false, "Don't show the session banner")
	cmd.Flags().StringVar(&ccmd.TFACode, "tfaCode", "", "Two Factor Authentication code")
	cmd.Flags().BoolVar(&ccmd.ClearToken, "clear", false, "Clear any existing tokens")
	cmd.Flags().StringVar(&ccmd.Format, "format", "", "External command format, e.g. json, yaml, toml")
	cmd.Flags().StringVar(&ccmd.OutputFormat, "output-format", "", "Output format")
	cmd.Flags().StringVar(&ccmd.Shell, "shell", "", "Shell type to return the environment variables")
	cmd.Flags().StringVar(&ccmd.LoginType, "loginType", "", "Login type preference, e.g. OAUTH2_INTERNAL, DEVICE (device flow), BROWSER (Authorization Code via browser), BASIC or CERTIFICATE (mTLS). OAUTH2 is a legacy alias for DEVICE. When set to BASIC, any existing token will be cleared")
	cmd.Flags().StringVar(&ccmd.Mode, "mode", "", "Session mode which controls which commands are allowed, e.g. dev, qual or prod")
	cmd.Flags().StringSliceVar(&ccmd.Secrets, "secrets", []string{}, "List of secrets to include as env variables when running an external command or extension provider. Only valid with from-cmd or from-extension")
	cmd.Flags().StringVar(&ccmd.Host, "host", "", "Cumulocity host. Only used with the 'interactive' provider")
	cmd.Flags().StringVar(&ccmd.SSOScopes, "sso-scopes", "", "SSO Scopes")
	cmd.Flags().StringVar(&ccmd.SSOAudience, "sso-audience", "", "SSO Audience")
	cmd.Flags().StringVar(&ccmd.BrowserCallbackURL, "browserCallback", "", "Custom redirect URI for the browser flow local callback server, e.g. http://127.0.0.1:8080/callback. Must match a URI registered in the SSO provider. Defaults to http://127.0.0.1:5001/callback")

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
			config.ProviderTypeExtension,
		),
		completion.WithValidateSet("format", "json", "yaml", "toml", "dotenv"),
		completion.WithValidateSet("loginType", logintype.OAuth2Internal, logintype.Basic, logintype.None, logintype.OAuth2, logintype.Device, logintype.Browser, logintype.Certificate),
		completion.WithValidateSet(
			"mode",
			config.GetSessionModeCompletionHelp()...,
		),
	)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	cmd.MarkFlagsMutuallyExclusive("from-file", "from-cmd", "from-stdin", "from-env", "from-prompt", "from-extension")
	_ = cmd.RegisterFlagCompletionFunc("from-extension", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		providers := extensions.ListSessionProviders(f.ExtensionManager().List(), nil)
		return extensions.SessionProviderCompletions(providers), cobra.ShellCompDirectiveNoFileComp
	})
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

	// Certificate paths for mTLS auth
	if v := strings.TrimSpace(os.Getenv("C8Y_CERTIFICATE")); v != "" {
		session.Certificate = v
	}
	if v := strings.TrimSpace(os.Getenv("C8Y_CERTIFICATE_KEY")); v != "" {
		session.CertificateKey = v
	}

	// Login type — explicit env var wins; auto-detect CERTIFICATE when cert paths are present
	if v := strings.TrimSpace(os.Getenv("C8Y_SETTINGS_LOGIN_TYPE")); v != "" {
		session.LoginType = v
	} else if session.Certificate != "" {
		session.LoginType = logintype.Certificate
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
		modeOptions := config.GetSessionModeCompletionHelp()
		mode, err := prompt.Select("Select mode", modeOptions, modeOptions[0])
		if err != nil {
			return nil, err
		}
		session.Mode = mode
	}

	session.SessionUri = "interactive://host"
	return session, nil
}

// providerFlags returns the flags which are passed to external and extension session providers
func (n *CmdLogin) providerFlags(cfg *config.Config) []string {
	flags := []string{}
	if n.ClearToken {
		flags = append(flags, "--clear")
	}
	if n.LoginType != "" {
		flags = append(flags, fmt.Sprintf("--loginType=%s", n.LoginType))
	}
	if n.BrowserCallbackURL != "" {
		flags = append(flags, fmt.Sprintf("--browserCallback=%s", n.BrowserCallbackURL))
	}
	if n.SSOScopes != "" {
		flags = append(flags, fmt.Sprintf("--sso-scopes=%s", n.SSOScopes))
	}
	if n.SSOAudience != "" {
		flags = append(flags, fmt.Sprintf("--sso-audience=%s", n.SSOAudience))
	}
	if cfg.Verbose() {
		flags = append(flags, "-v")
	}
	if cfg.Debug() {
		flags = append(flags, "--debug")
	}
	return flags
}

// readSecrets returns the values of the given secrets ("KEY=value"), from the environment if set,
// otherwise from the pin entry prompt. Secrets without a value are skipped
func readSecrets(cfg *config.Config, keys []string) []string {
	values := []string{}
	seen := map[string]bool{}
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		if v := os.Getenv(key); v != "" {
			values = append(values, fmt.Sprintf("%s=%s", key, v))
			continue
		}
		secret, secretErr := cfg.PromptSecret(key)
		if secretErr != nil {
			if !errors.Is(secretErr, prompt.ErrNoPrompter) {
				cfg.Logger.Warnf("Could not get secret. key=%s, err=%s", key, secretErr)
			} else {
				cfg.Logger.Debugf("Secret is not set and no pin entry is configured. key=%s", key)
			}
			continue
		}
		cfg.Logger.Debugf("Setting env variable secret for session provider. %s", key)
		values = append(values, fmt.Sprintf("%s=%s", key, secret))
	}
	return values
}

// runProvider runs a session provider command and parses the session which it writes to stdout.
// stdin and stderr are connected so the provider can interact with the user
func (n *CmdLogin) runProvider(argv []string, env []string, format string) (*c8ysession.CumulocitySession, error) {
	return n.runProviderWithStdin(argv, env, format, os.Stdin)
}

// runProviderWithStdin runs a session provider command with the given stdin (nil for none)
func (n *CmdLogin) runProviderWithStdin(argv []string, env []string, format string, stdin io.Reader) (*c8ysession.CumulocitySession, error) {
	cfg, err := n.factory.Config()
	if err != nil {
		return nil, err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Env = env
	cmd.Stdin = stdin
	cmd.Stderr = os.Stderr

	log.Infof("Executing session provider: %s", shellquote.Join(argv...))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	if format == "" {
		// Try to detect the format
		if jsonUtilities.IsJSONObject(output) {
			format = "json"
			log.Infof("Detected input format: %s", format)
		} else {
			format = "dotenv"
			log.Infof("Guessing input format: %s", format)
		}
	}

	if client, err := n.factory.ClientSession(); err == nil {
		if cfg.HideSensitive() {
			log.Infof("Parsing session provider output: %s", cfg.HideSensitiveInformation(&client, strings.TrimSpace(string(output))))
		} else {
			log.Infof("Parsing session provider output: %s", strings.TrimSpace(string(output)))
		}
	} else {
		log.Infof("Parsing session provider output")
	}
	return n.FromReader(bytes.NewReader(output), format)
}

func (n *CmdLogin) FromExternalProvider(args []string) (*c8ysession.CumulocitySession, error) {
	cfg, err := n.factory.Config()
	if err != nil {
		return nil, err
	}

	// add secrets when executing the environment in case
	// if the external command requires extra authentication.
	// Note: Secrets which are already set are part of the environment
	env := os.Environ()
	missingSecrets := []string{}
	for _, key := range n.Secrets {
		if os.Getenv(key) == "" {
			missingSecrets = append(missingSecrets, key)
		}
	}
	env = append(env, readSecrets(cfg, missingSecrets)...)

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
	providerCommand = append(providerCommand, n.providerFlags(cfg)...)
	providerCommand = append(providerCommand, args...)

	return n.runProvider(providerCommand, env, n.Format)
}

// extensionProviderEnv returns the environment of an extension session provider: the current
// environment without the variables of the active session, the extension variables, and the secrets
func extensionProviderEnv(cfg *config.Config, provider *extensions.InstalledSessionProvider, executable string, secrets []string) []string {
	unset := map[string]bool{}
	for _, key := range c8ysession.GetSessionEnvKeys() {
		unset[key] = true
	}
	// The secrets are added again below
	keys := append(append([]string{}, provider.Provider.Secrets...), secrets...)
	for _, key := range keys {
		delete(unset, key)
	}
	set := extensions.Variables(provider.Extension, executable)
	base := extensions.MergeEnv(os.Environ(), nil, unset)
	env := extensions.MergeEnv(base, set, nil)

	secretValues := readSecrets(cfg, keys)
	names := []string{}
	for _, item := range secretValues {
		name, _, _ := strings.Cut(item, "=")
		names = append(names, name)
	}
	cfg.Logger.Infof("Passing secrets to session provider. provider=%s, secrets=%v", provider.Ref(), names)
	secretSet := map[string]string{}
	for _, item := range secretValues {
		name, value, _ := strings.Cut(item, "=")
		secretSet[name] = value
	}
	return extensions.MergeEnv(env, secretSet, nil)
}

// resolveExtensionProvider finds the extension session provider, and checks that its binary exists
func resolveExtensionProvider(f *cmdutil.Factory, cfg *config.Config, ref string) (*extensions.InstalledSessionProvider, string, error) {
	if strings.TrimSpace(ref) == "" {
		return nil, "", cmderrors.NewUserError("no extension session provider is set. Use --from-extension <provider>, or set it using 'c8y settings update session.provider.extension <provider>'")
	}

	providers := extensions.ListSessionProviders(f.ExtensionManager().List(), func(ext extensions.Extension, msg string) {
		cfg.Logger.Warnf("%s. extension=%s", msg, ext.Name())
	})
	provider, err := extensions.ResolveSessionProvider(providers, ref)
	if err != nil {
		return nil, "", cmderrors.NewUserError(err)
	}

	exe := provider.Extension.Executable()
	if _, err := os.Stat(exe); err != nil {
		return nil, "", cmderrors.NewUserError(fmt.Sprintf("the binary of extension %s was not found (%s). Reinstall or update the extension using 'c8y extensions update %s'", provider.Extension.Name(), exe, provider.Extension.Name()))
	}
	return provider, exe, nil
}

// FromExtensionProvider gets the session from a session provider of an installed extension
func (n *CmdLogin) FromExtensionProvider(ref string, args []string) (*c8ysession.CumulocitySession, error) {
	cfg, err := n.factory.Config()
	if err != nil {
		return nil, err
	}
	provider, exe, err := resolveExtensionProvider(n.factory, cfg, ref)
	if err != nil {
		return nil, err
	}

	format := n.Format
	if format == "" {
		format = provider.Provider.OutputFormat()
	}
	env := extensionProviderEnv(cfg, provider, n.factory.Executable, n.Secrets)

	argv := []string{exe}
	if n.Select != "" {
		// Select the session without user interaction
		if !provider.Provider.CanList() {
			return nil, cmderrors.NewUserError(fmt.Sprintf("session provider %s does not support --select, as it does not declare the list and select commands", provider.Ref()))
		}
		argv = append(argv, provider.Provider.Select...)
		argv = append(argv, n.providerFlags(cfg)...)
		argv = append(argv, args...)
		argv = append(argv, n.Select)
		return n.runProviderWithStdin(argv, env, format, nil)
	}

	argv = append(argv, provider.Provider.Exec...)
	argv = append(argv, n.providerFlags(cfg)...)
	argv = append(argv, args...)
	return n.runProvider(argv, env, format)
}

// DefaultListTimeout is the maximum duration of a session provider's list command
const DefaultListTimeout = 30 * time.Second

// ListExtensionSessions runs the list command of an extension session provider, and returns the
// provider and its sessions (without secrets). The provider gets the same environment and
// secrets as when logging in, but no stdin, so it can't interact with the user. args are passed
// to the list command, and secrets are additional secrets to pass to the provider
func ListExtensionSessions(ctx context.Context, f *cmdutil.Factory, ref string, args []string, secrets []string) (*extensions.InstalledSessionProvider, []extensions.ProviderSession, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, nil, err
	}
	provider, exe, err := resolveExtensionProvider(f, cfg, ref)
	if err != nil {
		return nil, nil, err
	}
	if !provider.Provider.CanList() {
		return provider, nil, cmderrors.NewUserError(fmt.Sprintf("session provider %s can't list its sessions, as it does not declare the list and select commands", provider.Ref()))
	}

	argv := append([]string{exe}, provider.Provider.List...)
	argv = append(argv, args...)

	if ctx == nil {
		ctx = context.Background()
	}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultListTimeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Env = extensionProviderEnv(cfg, provider, f.Executable, secrets)
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	cfg.Logger.Infof("Listing sessions of session provider: %s", shellquote.Join(argv...))
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return provider, nil, fmt.Errorf("listing the sessions of %s timed out", provider.Ref())
		}
		// cobra based providers prefix the error with "Error: "
		if reason := strings.TrimPrefix(lastLine(stderr.String()), "Error: "); reason != "" {
			return provider, nil, fmt.Errorf("%s", reason)
		}
		return provider, nil, err
	}
	sessions, warnings, err := extensions.ParseProviderSessions(output)
	for _, w := range warnings {
		cfg.Logger.Warnf("%s. provider=%s", w, provider.Ref())
	}
	return provider, sessions, err
}

// lastLine returns the last non-empty line of the text
func lastLine(text string) string {
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			return line
		}
	}
	return ""
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
		SessionUri:         getValue("sessionUri"),
		Path:               getValue("path"),
		Username:           getValue("username"),
		Password:           getValue("password"),
		Tenant:             getValue("tenant"),
		Token:              getValue("token"),
		TOTP:               getValue("totp"),
		Mode:               getValue("mode"),
		LoginType:          getValue("loginType"),
		Version:            getValue("version"),
		Certificate:        getValue("certificate"),
		CertificateKey:     getValue("certificateKey", "certificate_key"),
		BrowserCallbackURL: getValue("settings.sso.browserCallbackUrl", "browserCallbackUrl"),
	}

	// Get if value is defined before accessing it
	if v.GetString("authorized") != "" {
		session.SetAuthorized(v.GetBool("authorized"))
	}

	session.SetHost(getValue("host"))

	// Auto-detect CERTIFICATE login type when cert paths are present and no loginType is set
	if session.LoginType == "" && session.Certificate != "" {
		session.LoginType = logintype.Certificate
	}

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

	if cmd.Flags().Changed("from-extension") {
		// Explicit extension provider overrides the configured provider type
		n.Provider = config.ProviderTypeExtension
	}

	if !oneHasChanged(cmd, "from-cmd", "from-file", "from-env", "from-stdin", "from-prompt", "from-extension") {

		// Set defaults from config if values from flags aren't provided
		if !cmd.Flags().Changed("provider") {
			n.Provider = cfg.SessionProvider()
			cfg.Logger.Debugf("Using session provider from configuration. type=%s", n.Provider)
		}

		n.Exec = cfg.SessionProviderCommand()
		n.FromExtension = cfg.SessionProviderExtension()
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

	if n.Select != "" && !strings.EqualFold(n.Provider, config.ProviderTypeExtension) {
		return cmderrors.NewUserError("--select is only valid with an extension session provider (--from-extension)")
	}

	var session *c8ysession.CumulocitySession

	switch strings.ToLower(n.Provider) {
	case config.ProviderTypeInteractive:
		session, err = n.FromInteractive(cmd)
	case config.ProviderTypeExternal:
		session, err = n.FromExternalProvider(args)
	case config.ProviderTypeExtension:
		session, err = n.FromExtensionProvider(n.FromExtension, args)
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

	// Session files can contain an encrypted password and token
	if err := n.decryptSession(cfg, session); err != nil {
		return err
	}

	// Fail early if the domain is not set
	if session.GetDomain() == "" {
		return cmderrors.NewUserError("invalid session. host is empty")
	}

	if sessionContents, err := json.Marshal(session); err == nil {

		if client, err := n.factory.ClientSession(); err == nil {
			if cfg.HideSensitive() {
				log.Infof("Received session from external source:\n%s\n", cfg.HideSensitiveInformation(&client, string(sessionContents)))
			} else {
				cfg.Logger.Infof("Received session from external source:\n%s\n", sessionContents)
			}
		} else {
			log.Infof("Received session from external source")
		}
	}

	if session.SessionUri != "" {
		cfg.SetSessionFile(session.SessionUri)
	} else {
		cfg.SetSessionFile(session.Path)
	}

	// The credentials of the session, and the authorization which is used for requests
	auth := authentication.AuthOptions{
		Tenant:   session.Tenant,
		Username: session.Username,
		Password: session.Password,
	}
	authType := authentication.AuthTypeUnset
	if session.Token != "" {
		auth.Token = session.Token
		authType = authentication.AuthTypeBearer
	} else if session.Username != "" && session.Password != "" {
		authType = authentication.AuthTypeBasic
	}
	baseURL, _ := url.Parse(apiv2.FormatBaseURL(session.Host))
	client := config.NewClientSession(baseURL, auth, "")
	client.AuthType = authType

	if !session.IsAuthorized() {
		loginType := strings.ToUpper(cfg.GetLoginTypeWithDefault())
		if session.LoginType != "" {
			loginType = strings.ToUpper(session.LoginType)
		}
		if n.LoginType != "" {
			loginType = strings.ToUpper(n.LoginType)
		}
		log.Infof("User flag login type: %s", loginType)

		// Set default auth mode based on login type
		switch loginType {
		case logintype.OAuth2Internal, logintype.OAuth2, logintype.Device, logintype.Browser, logintype.Certificate:
			authType = authentication.AuthTypeBearer
		case logintype.Basic:
			authType = authentication.AuthTypeBasic
		case logintype.None:
			authType = authentication.AuthTypeNone
		}

		// SSO providers are in control of the token, so it should just be used as is
		if !n.ClearToken && c8ysession.ShouldReuseToken(cfg, log, session.Token, loginType) {
			auth.Token = session.Token
			authType = authentication.AuthTypeBearer
		} else {
			auth.Token = ""
		}
		auth.AuthType = []authentication.AuthType{authType}

		c8ysession.ClearProcessEnvironment()

		// The login type flows which are provided by the SDK use the connection settings. The other
		// login types have always accepted self-signed certificates (e.g. for edge devices).
		insecure := true
		if loginType == logintype.Browser || loginType == logintype.Certificate || loginType == logintype.Device {
			insecure = cfg.SkipSSLVerify()
		}
		handler, err := c8ylogin.NewLoginHandlerV2(
			n.factory.IOStreams,
			apiv2.ClientOptions{
				BaseURL:            session.Host,
				Auth:               auth,
				InsecureSkipVerify: insecure,
			},
			cmd.ErrOrStderr(),
			func() {},
		)
		if err != nil {
			return err
		}
		handler.LoginType = loginType
		handler.CertificatePath = session.Certificate
		handler.CertificateKeyPath = session.CertificateKey
		if loginType != "" {
			handler.AllowedLoginTypes = strings.Split(loginType, ",")
		}
		handler.SSO.DiscoveryURL = cfg.SSODiscoveryUrl()
		if len(n.SSOScopes) > 0 {
			handler.SSO.Scopes = strings.Split(n.SSOScopes, " ")
		} else {
			handler.SSO.Scopes = cfg.SSOScopes()
		}
		handler.SSO.Audience = n.SSOAudience
		handler.BrowserCallbackURL = n.BrowserCallbackURL
		if handler.BrowserCallbackURL == "" {
			handler.BrowserCallbackURL = session.BrowserCallbackURL
		}
		if handler.BrowserCallbackURL == "" {
			handler.BrowserCallbackURL = cfg.BrowserCallbackURL()
		}
		handler.TOTPSecret = session.TOTP
		handler.TOTPCode = cfg.TOTPCode
		if handler.TOTPCode == "" {
			handler.TOTPCode = n.TFACode
		}
		handler.SetLogger(log)
		log.Infof("User preference for login type: %s", handler.LoginType)
		if err := handler.Run(); err != nil {
			return err
		}
		client = handler.Session()

		session.Username = client.Username
		session.Host = client.URL

		// Use the handler selected login type
		session.LoginType = handler.LoginType

		if client.Version != "" {
			session.Version = client.Version
		}

		if client.Tenant != "" {
			session.Tenant = client.Tenant
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
		c8ysession.PrintSessionInfo(n.SubCommand.GetCommand().ErrOrStderr(), &client, cfg, *session)
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
	return c8ysession.WriteOutput(n.GetCommand().OutOrStdout(), &client, cfg, session, outputFormat)
}

// decryptSession decrypts the password and token if they are encrypted (e.g. when read
// from a session file), prompting for the passphrase if it isn't already set via C8Y_PASSPHRASE
func (n *CmdLogin) decryptSession(cfg *config.Config, session *c8ysession.CumulocitySession) error {
	encryptedValue := ""
	for _, value := range []string{session.Password, session.Token} {
		if cfg.SecureData.IsEncrypted(value) == 1 {
			encryptedValue = value
			break
		}
	}
	if encryptedValue == "" {
		return nil
	}

	// Validate the passphrase against the encrypted value itself, so no key file is needed
	passphrase, err := cfg.CheckEncryption(encryptedValue)
	if err != nil {
		return fmt.Errorf("could not decrypt the session. %w", err)
	}
	if passphrase == "" || passphrase == "null" {
		return cmderrors.NewUserError("could not decrypt the session. passphrase can not be empty")
	}

	if session.Password, err = cfg.SecureData.TryDecryptString(session.Password, passphrase); err != nil {
		return fmt.Errorf("could not decrypt the session password. %w", err)
	}
	if session.Token, err = cfg.SecureData.TryDecryptString(session.Token, passphrase); err != nil {
		return fmt.Errorf("could not decrypt the session token. %w", err)
	}

	if n.factory.IOStreams != nil {
		fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s Passphrase OK\n", n.factory.IOStreams.ColorScheme().SuccessIcon())
	}
	return nil
}
