package login

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/kballard/go-shellquote"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ylogin"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
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
	Format   string
	Provider string

	// Login options
	LoginType string

	// Output options
	Shell        string
	OutputFormat string

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

var (
	ProviderTypeAuto     = "auto"
	ProviderTypeFile     = "file"
	ProviderTypeEnv      = "env"
	ProviderTypeExternal = "external"
	ProviderTypeStdin    = "stdin"
)

func NewCmdLogin(f *cmdutil.Factory) *CmdLogin {
	ccmd := &CmdLogin{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "login to Cumulocity IoT and return environment variables (including a token)",
		Long:  `Set a session, login and test the session and get either OAuth2 token, or using two factor authentication`,
		Example: heredoc.Doc(`
			$ eval "$( c8y-session-bitwarden | c8y session login --from-stdin )"
			Set a session interactively

			$ eval "$( c8y sessions login --exec "c8y-session-bitwarden list --folder c8y" )"
			Set a session but only include session matching company AND dev

			$ eval "$( c8y sessions login --from-file .env )"
			Set a session from a dotenv file

			$ eval "$( c8y sessions login --from-env )"
			Set a session from existing environment variables
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.Provider, "provider", "", "Session provider which returns the session to use")
	cmd.Flags().StringVar(&ccmd.File, "from-file", "", "Read session from a file")
	cmd.Flags().StringVar(&ccmd.Exec, "from-cmd", "", "External command to execute to get the log in details")
	cmd.Flags().BoolVar(&ccmd.Env, "from-env", false, "Read from environment variables")
	cmd.Flags().BoolVar(&ccmd.Stdin, "from-stdin", false, "Read from standard input")
	cmd.Flags().StringVar(&ccmd.Format, "format", "", "External command format, e.g. json, yaml, toml")
	cmd.Flags().StringVar(&ccmd.OutputFormat, "output-format", "", "Output format")
	cmd.Flags().StringVar(&ccmd.Shell, "shell", "", "Shell type to return the environment variables")
	cmd.Flags().StringVar(&ccmd.LoginType, "loginType", "", "Login type preference, e.g. OAUTH2_INTERNAL or BASIC. When set to BASIC, any existing token will be cleared")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("shell", "auto", "bash", "zsh", "fish", "powershell"),
		completion.WithValidateSet("output-format", "json", "dotenv"),
		completion.WithValidateSet("provider", ProviderTypeFile, ProviderTypeStdin, ProviderTypeEnv, ProviderTypeExternal, ProviderTypeAuto),
		completion.WithValidateSet("format", "json", "yaml", "toml", "dotenv"),
		completion.WithValidateSet("loginType", c8y.AuthMethodOAuth2Internal, c8y.AuthMethodBasic),
	)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	cmd.MarkFlagsMutuallyExclusive("from-file", "from-cmd", "from-stdin", "from-env")
	cmd.MarkFlagsMutuallyExclusive("output-format", "shell")

	return ccmd
}

func (n *CmdLogin) FromStdin(format string, args []string) (*c8ysession.CumulocitySession, error) {
	return n.FromReader(bufio.NewReader(os.Stdin), format)
}

func (n *CmdLogin) FromEnv() (*c8ysession.CumulocitySession, error) {
	session := &c8ysession.CumulocitySession{
		Tenant:   os.Getenv("C8Y_TENANT"),
		Password: os.Getenv("C8Y_PASSWORD"),
		Token:    os.Getenv("C8Y_TOKEN"),
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

	providerCmd := strings.TrimSpace(n.Exec)
	if providerCmd == "" {
		providerCmd = strings.TrimSpace(cfg.GetString("settings.session.providerCmd"))
	}
	if providerCmd == "" {
		return nil, fmt.Errorf("provider is not set")
	}

	cmdArgs, err := shellquote.Split(providerCmd)
	if err != nil {
		return nil, err
	}
	if len(cmdArgs) == 0 {
		return nil, fmt.Errorf("executable is empty")
	}
	cmdExec := cmdArgs[0]
	cmd := exec.Command(cmdExec, slices.Concat(cmdArgs[1:], args)...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	log.Infof("Executing session provider: %s", providerCmd)
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
	session := &c8ysession.CumulocitySession{
		SessionUri: v.GetString("sessionUri"),
		Path:       v.GetString("path"),
		Username:   v.GetString("username"),
		Password:   v.GetString("password"),
		Tenant:     v.GetString("tenant"),
		Token:      v.GetString("token"),
		TOTP:       v.GetString("totp"),
	}
	session.SetHost(v.GetString("host"))
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

	return n.FromViper(v)
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

func (n *CmdLogin) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}

	if n.Provider == "" {
		n.Provider = cfg.GetString("settings.session.provider")
		if n.Provider == "" {
			n.Provider = ProviderTypeAuto
		}
	}

	if strings.EqualFold(n.Provider, ProviderTypeAuto) {
		//
		// Try guessing a sensible default
		//
		if n.File != "" {
			n.Provider = ProviderTypeFile
		} else if n.Stdin {
			n.Provider = ProviderTypeStdin
		} else if n.Env {
			n.Provider = ProviderTypeEnv
		} else if n.Exec != "" {
			n.Provider = ProviderTypeExternal
		} else if os.Getenv("CI") != "" {
			// CI environment and generally env variables are used here
			n.Provider = ProviderTypeEnv
		} else if n.factory.IOStreams.HasStdin() {
			n.Provider = ProviderTypeStdin
		}
	}

	var session *c8ysession.CumulocitySession

	switch strings.ToLower(n.Provider) {
	case ProviderTypeExternal:
		session, err = n.FromExternalProvider(args)
	case ProviderTypeEnv:
		session, err = n.FromEnv()
	case ProviderTypeStdin:
		if !n.factory.IOStreams.HasStdin() {
			err = fmt.Errorf("no stdin detected")
		} else {
			session, err = n.FromStdin(n.Format, args)
		}
	case ProviderTypeFile:
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

	client := c8y.NewClient(nil, session.Host, session.Tenant, session.Username, session.Password, true)
	client.SetToken(session.Token)

	c8ysession.ClearProcessEnvironment()

	handler := c8ylogin.NewLoginHandler(client, cmd.ErrOrStderr(), func() {})
	handler.Interactive = true
	handler.LoginType = strings.ToUpper(cfg.GetLoginType())
	if n.LoginType != "" {
		handler.LoginType = strings.ToUpper(n.LoginType)
	}

	log.Infof("User preference for login type: %s", handler.LoginType)
	handler.TFACode = session.TOTP
	handler.SetLogger(log)
	err = handler.Run()
	if err != nil {
		return err
	}

	session.Token = client.Token
	if client.TenantName != "" {
		session.Tenant = client.TenantName
	}
	session.Version = client.Version
	session.Username = handler.C8Yclient.Username
	session.Host = handler.C8Yclient.BaseURL.Host
	session.Path = cfg.GetSessionFile()

	// Write session details to stderr (for humans)
	c8ysession.PrintSessionInfo(n.SubCommand.GetCommand().ErrOrStderr(), client, cfg, *session)

	outputFormat := n.OutputFormat
	if outputFormat == "" {
		if n.Shell == "" && !n.factory.IOStreams.IsStdoutTTY() {
			n.Shell = "auto"
		}
		if strings.EqualFold(n.Shell, "auto") {
			n.Shell = shell.DetectShell("bash")
		}
		outputFormat = n.Shell
	}

	// Write session details to stdout (for machines)
	return c8ysession.WriteOutput(n.GetCommand().OutOrStdout(), client, cfg, session, outputFormat)
}
