package create

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileutilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logintype"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/prompt"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCumulocitySessionFromFile(filePath string, log *logger.Logger, cfg *config.Config) (*c8ysession.CumulocitySession, error) {
	session := &c8ysession.CumulocitySession{
		Config: cfg,
		Logger: log,
	}

	sessionConfig := viper.New()
	sessionConfig.SetConfigFile(filePath)
	if !config.SupportsFileExtension(filePath) {
		sessionConfig.SetConfigType("json")
	}

	if err := sessionConfig.ReadInConfig(); err != nil {
		return nil, err
	}
	session.Schema = sessionConfig.GetString("$schema")
	session.Name = sessionConfig.GetString("name")
	session.Description = sessionConfig.GetString("description")
	session.Host = sessionConfig.GetString("host")
	session.Tenant = sessionConfig.GetString("tenant")
	session.Username = sessionConfig.GetString("username")
	session.Password = sessionConfig.GetString("password")
	session.Token = sessionConfig.GetString("token")
	session.UseTenantPrefix = sessionConfig.GetBool("useTenantPrefix")
	session.MicroserviceAliases = sessionConfig.GetStringMapString("microserviceAliases")

	session.Path = filePath
	session.Extension = filepath.Ext(filePath)[1:]

	basename := filepath.Base(filePath)
	extension := filepath.Ext(basename)
	session.Name = strings.TrimSuffix(basename, extension)
	return session, nil
}

type CmdCreate struct {
	host               string
	username           string
	password           string
	token              string
	description        string
	name               string
	tenant             string
	sessionMode        string
	loginType          string
	noTenantPrefix     bool
	noStorage          bool
	encrypt            bool
	browserCallbackURL string
	allowInsecure      bool
	prompt             bool

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdCreate(f *cmdutil.Factory) *CmdCreate {
	ccmd := &CmdCreate{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create session",
		Long:  `Create a new Cumulocity session`,
		Example: heredoc.Doc(`
$ c8y sessions create --mode dev --host "https://mytenant.eu-latest.cumulocity.com"
Example 1: Create a DEV new session. Prompt for username and password

$ c8y sessions create \
    --mode qual \
	--host "https://mytenant.eu-latest.cumulocity.com"
	--username "myUser@me.com"
Create a new QA (QUAL) session prompting for password

$ c8y sessions create --mode prod --host "https://mytenant.eu-latest.cumulocity.com" --noStorage
Create a new production session where only only GET commands are enabled (with no password storage)

$ c8y sessions create --mode prod --host "https://localhost:443" --insecure
Create a session which points to a local api endpoint (most like an Cumulocity Edge instance)

$ c8y sessions create --mode dev --host example.cumulocity.com --loginType BROWSER
Create a session which uses SSO / OAUTH2 using Authorization Flow (via a local web browser)
Note: Requires the "Redirect to the user interface application" to be enabled in Cumulocity

c8y sessions create --mode dev --host example.cumulocity.com --loginType BROWSER --browserCallback localhost:8008/mycallback
Create a session which uses SSO / OAUTH2 using Authorization Flow (via a local web browser)
and define an explicit redirect/callback URI which is whitelisted in the SSO providers configuration
Note: Requires the "Redirect to the user interface application" to be enabled in Cumulocity

$ c8y sessions create --mode dev --host example.cumulocity.com --loginType DEVICE
Create a session with SSO / OAUTH2 Device Flow (RFC 8628)
		`),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.InheritedFlags().Changed("examples") {
				return cmd.Root().PersistentPreRunE(cmd, args)
			}
			return ccmd.promptArgs(cmd, args)
		},
		Args: cobra.NoArgs,
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.host, "host", "", "Host. .e.g. test.cumulocity.com. (required)")
	cmd.Flags().StringVar(&ccmd.username, "username", "", "Username (without tenant). (required)")
	cmd.Flags().StringVar(&ccmd.password, "password", "", "Password. If left blank then you will be prompted for the password")
	cmd.Flags().StringVar(&ccmd.token, "token", "", "Token")
	cmd.Flags().StringVar(&ccmd.tenant, "tenant", "", "Tenant ID")
	cmd.Flags().StringVar(&ccmd.description, "description", "", "Description about the session")
	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of the session")
	cmd.Flags().StringVar(&ccmd.sessionMode, "type", "", "Session type. List of predefined session types (deprecated)")
	cmd.Flags().StringVar(&ccmd.sessionMode, "mode", "", "Session mode which controls which commands are enabled by default")
	cmd.Flags().StringVar(&ccmd.loginType, "loginType", "", "Login Type, e.g. BASIC, OAUTH2_INTERNAL, OAUTH2, BROWSER, DEVICE, CERTIFICATE, NONE")
	cmd.Flags().BoolVar(&ccmd.noTenantPrefix, "noTenantPrefix", false, "Don't use tenant name as a prefix to the user name when using Basic Authentication. Defaults to false")
	cmd.Flags().BoolVar(&ccmd.noStorage, "noStorage", false, "Don't store any passwords or tokens in the session file")
	cmd.Flags().BoolVar(&ccmd.encrypt, "encrypt", false, "Encrypt passwords and tokens (occurs when logging in)")
	cmd.Flags().BoolVar(&ccmd.allowInsecure, "allowInsecure", false, "Allow insecure connection (e.g. when using self-signed certificates)")
	cmd.Flags().BoolVar(&ccmd.prompt, "prompt", false, "Force prompting of missing information")
	cmd.Flags().StringVar(&ccmd.browserCallbackURL, "browserCallback", "", "Custom redirect URI for the browser authorization code flow, e.g. http://127.0.0.1:8080/callback")

	// Required flags
	completion.WithOptions(cmd,
		completion.WithLazyRequired("type"),
		completion.WithValidateSet(
			"type",
			config.GetSessionModeCompletionHelp()...,
		),
		completion.WithValidateSet(
			"mode",
			config.GetSessionModeCompletionHelp()...,
		),
		completion.WithValidateSet(
			"loginType",
			c8y.LoginTypeBasic,
			c8y.LoginTypeOAuth2Internal,
			c8y.LoginTypeOAuth2,
			logintype.Browser,
			logintype.Device,
			logintype.Certificate,
			c8y.LoginTypeNone,
		),
	)

	flags.MarkDeprecated(cmd, "type", "please use 'mode' instead")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdCreate) promptArgs(cmd *cobra.Command, args []string) error {
	if n.prompt {
		n.factory.IOStreams.SetStdoutTTY(true)
		n.factory.IOStreams.SetStdinTTY(true)
	}
	if !n.factory.IOStreams.CanPrompt() {
		return nil
	}
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	c8y.Logger = log
	prompter := prompt.NewPrompt(log)

	if !cmd.Flags().Changed("host") {
		v, err := prompter.Input("Enter host", "", true, false)

		if err != nil {
			return err
		}
		n.host = strings.TrimSpace(v)
	}

	if !cmd.Flags().Changed("loginType") {
		localClient := c8y.NewClientFromOptions(nil, c8y.ClientOptions{
			BaseURL: n.host,
		})
		loginOptions, _, loginOptionsErr := localClient.Tenant.GetLoginOptions(context.Background())
		if loginOptionsErr != nil {
			return loginOptionsErr
		}

		loginOptionsForUsers := make([]string, 0, len(loginOptions.LoginOptions))
		loginOptionsForUsers = append(loginOptionsForUsers, "auto\tUse tenant default")
		hasOAuth2 := false
		for _, option := range loginOptions.LoginOptions {
			loginOptionsForUsers = append(loginOptionsForUsers, option.Type)
			if strings.EqualFold(option.Type, logintype.OAuth2) {
				hasOAuth2 = true
			}
		}
		// BROWSER and DEVICE are CLI-side flows built on top of an external
		// OAUTH2 provider; add them whenever the tenant supports OAUTH2.
		if hasOAuth2 {
			loginOptionsForUsers = append(loginOptionsForUsers,
				logintype.Browser+"\tOAUTH2 via local browser (authorization code flow)",
				logintype.Device+"\tOAUTH2 device flow",
			)
		}
		// CERTIFICATE is always available as it uses mTLS, not a tenant login option.
		loginOptionsForUsers = append(loginOptionsForUsers, logintype.Certificate+"\tCertificate-based (mTLS)")

		selectedLoginOption, err := prompt.Select("Select login type", loginOptionsForUsers, loginOptionsForUsers[0])
		if err != nil {
			return err
		}
		// TODO: Add formal type for auto
		if selectedLoginOption != "auto" {
			n.loginType = selectedLoginOption
		}
	}

	loginTypeRequiresPasswords := requiresUsername(n.loginType)

	if !cmd.Flags().Changed("username") && loginTypeRequiresPasswords {
		v, err := prompter.Username("Enter username", " "+cfg.GetDefaultUsername())

		if err != nil {
			return err
		}
		n.username = strings.TrimSpace(v)
	}

	if !n.noStorage && !cmd.Flags().Changed("password") && loginTypeRequiresPasswords {
		password, err := prompter.Password("Enter c8y password", "")
		if err != nil {
			return err
		}
		n.password = password
	}

	if !cmd.Flags().Changed("type") && !cmd.Flags().Changed("mode") {
		modeOptions := config.GetSessionModeCompletionHelp()
		mode, err := prompt.Select("Select mode", modeOptions, modeOptions[0])
		if err != nil {
			return err
		}
		n.sessionMode = mode
	}

	return nil
}

func requiresUsername(loginType string) bool {
	switch strings.ToUpper(loginType) {
	case c8y.LoginTypeNone, c8y.LoginTypeOAuth2, logintype.Browser, logintype.Device, logintype.Certificate:
		return false
	}
	return true
}

func (n *CmdCreate) RunE(cmd *cobra.Command, args []string) error {
	// Validate required parameters here as the user could have entered them
	// via the prompt
	if n.host == "" {
		return &flags.ParameterError{
			Name: "host",
			Err:  flags.ErrParameterMissing,
		}
	}

	if n.username == "" && requiresUsername(n.loginType) {
		return &flags.ParameterError{
			Name: "username",
			Err:  flags.ErrParameterMissing,
		}
	}

	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	session := &c8ysession.CumulocitySession{
		Schema:          "https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/v2/tools/schema/session.schema.json",
		Host:            n.host,
		Tenant:          n.tenant,
		Username:        n.username,
		Description:     n.description,
		UseTenantPrefix: !n.noTenantPrefix,
		Config:          cfg,
		Logger:          log,
	}

	session.MicroserviceAliases = make(map[string]string)

	settings := &config.CommandSettings{}
	settings.ActivityLog = &config.ActivityLogSettings{
		Enabled: settings.Bool(true),
	}

	if n.noStorage {
		settings.Storage = &config.StorageSettings{
			StorePassword: settings.Bool(false),
			StoreToken:    settings.Bool(false),
		}
	}

	if n.loginType != "" || n.browserCallbackURL != "" {
		if settings.Login == nil {
			settings.Login = &config.LoginSettings{}
		}
		if n.loginType != "" {
			settings.Login.Type = n.loginType
		}
	}

	if n.browserCallbackURL != "" {
		if settings.SSO == nil {
			settings.SSO = &config.SSOSettings{}
		}
		settings.SSO.BrowserCallbackURL = n.browserCallbackURL
	}

	switch n.sessionMode {
	case "ci":
		settings.Session = &config.SessionSettings{
			Confirmation: "PUT POST DELETE",
			Mode:         config.SessionModeCI.String(),
		}
	case "dev":
		settings.Session = &config.SessionSettings{
			Confirmation: "PUT POST DELETE",
			Mode:         config.SessionModeDev.String(),
		}
	case "qual":
		settings.Session = &config.SessionSettings{
			Confirmation: "PUT POST DELETE",
			Mode:         config.SessionModeQual.String(),
		}
	case "prod":
		settings.Session = &config.SessionSettings{
			Confirmation: "PUT POST DELETE",
			Mode:         config.SessionModeProduction.String(),
		}
	}

	if n.encrypt {
		settings.Encryption = &config.EncryptionSettings{
			Enabled:         settings.Bool(true),
			CachePassphrase: settings.Bool(true),
		}
	}

	if cmd.Flags().Changed("allowInsecure") {
		if settings.Defaults == nil {
			settings.Defaults = make(map[string]interface{})
		}
		settings.Defaults[config.GetSettingsNameWithoutPrefix(config.SettingsDefaultsInsecure)] = n.allowInsecure
	}

	session.Settings = settings

	if !n.noStorage {
		session.SetPassword(n.password)
		session.SetToken(n.token)
	}

	// session name (default to host and username)
	hostname := "c8y"
	if u, err := url.Parse(session.GetHost()); err == nil {
		// Don't include port number by default as it causes problems with paths across different OS's
		hostname = u.Hostname()
	}

	sessionName := hostname
	if session.Username != "" {
		sessionName += "-" + session.Username
	}

	if v, err := cmd.Flags().GetString("name"); err == nil && v != "" {
		sessionName = v
	}

	outputDir := cfg.GetSessionHomeDir()
	outputFile := n.formatFilename(sessionName)

	if err := n.writeSessionFile(outputDir, outputFile, *session); err != nil {
		return err
	}

	fmt.Println(path.Join(outputDir, outputFile))
	return nil
}

func (n *CmdCreate) formatFilename(name string) string {
	// Remove any characters that don't match the allowed chars
	reg := regexp.MustCompile(`[^A-Za-z0-9_#(). !,;'@%=&-]`)
	name = reg.ReplaceAllStringFunc(name, func(s string) string {
		// Replace invalid char with an empty string
		return ""
	})

	if !strings.HasSuffix(name, ".json") {
		name = fmt.Sprintf("%s.json", name)
	}
	return name
}

func (n *CmdCreate) writeSessionFile(outputDir, outputFile string, session c8ysession.CumulocitySession) error {
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(session, "", "  ")

	if err != nil {
		return errors.Wrap(err, "failed to convert session to json")
	}

	outputPath := path.Join(outputDir, outputFile)

	if outputDir != "" {
		if err := fileutilities.CreateDirs(outputDir); err != nil {
			log.Errorf("failed to create folder. folder=%s, err=%s", outputDir, err)
			return err
		}
	}
	log.Debugf("output file: %s", outputPath)

	if err := os.WriteFile(path.Join(outputDir, outputFile), data, 0600); err != nil {
		return errors.Wrap(err, "failed to write to file")
	}
	return nil
}
