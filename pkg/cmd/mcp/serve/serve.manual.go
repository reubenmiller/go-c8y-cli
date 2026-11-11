package serve

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mcpserver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CmdServe struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	policyFile    string
	mode          string
	allow         []string
	deny          []string
	denyPaths     []string
	readCommands  []string
	approval      string
	timeout       time.Duration
	maxTimeout    time.Duration
	maxOutput     int
	noDefaultDeny bool
	noNotify      bool

	allowDeviceShell   bool
	confirmDeviceShell bool
	stdio              bool
}

func NewCmdServe(f *cmdutil.Factory) *CmdServe {
	ccmd := &CmdServe{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start an MCP server for AI agents",
		Long: heredoc.Doc(`
			Start a Model Context Protocol (MCP) server which gives AI agents controlled access to a
			Cumulocity tenant using c8y commands.

			The MCP client starts the server and communicates with it over stdin/stdout, so register the
			command in the MCP client once (see the examples) rather than running it in a terminal.

			Sessions are selected and unlocked in the browser, as the server has no terminal:

			* Without --session, the first tool call opens a session picker in the browser. Selecting a
			  session logs in to it straight away and shows the result, so you can retry or pick another
			  session. The tool call waits until the login succeeds.
			* Encrypted sessions are unlocked with the pin entry command (settings.pinEntry, e.g.
			  "c8y settings update pinEntry 'touchie get'"), or else with the passphrase entered on the page.
			* If the login needs a two-factor authentication code, the page asks for it, unless the session
			  contains a TOTP secret.
			* Run the server's "switch_session" prompt (in Claude Code "/mcp__<server>__switch_session") to
			  select another session. Only the user can run it, not the agent.
			* The page only listens on 127.0.0.1 while it is open, and can only be opened with the one-time
			  link which the server opens in the browser. The link is never shown to the agent.
			* The passphrase and code are passed to the login via a pipe, and aren't kept. When the token
			  expires, the page opens again if they are needed.

			For automation and CI, set C8Y_HOST, C8Y_TENANT, C8Y_USER and C8Y_PASSWORD or C8Y_TOKEN in the MCP
			client configuration instead.

			You control what the agent can do:

			* Session: selected by you in the browser. The agent can't switch the session, host or
			  credentials. Commands only receive the token (or the password for basic auth sessions), never
			  the passphrase.
			* Mode (prod, qual, dev): the maximum mode the agent can use. By default, the agent uses the mode
			  of the session (settings.session.mode), and prod (read only) if the session has none. --mode can
			  only narrow the session's own mode, never widen it, also after switching the session. Read-only
			  commands always run in prod (read-only) mode.
			* Allow and deny lists: which commands the agent can use, e.g. "devices", "inventory *", "* delete".
			* Local files: commands can't read files in the session home folder or the session file (e.g. with
			  --template, --data or --file), so the agent can't read secrets from session files. Use --denyPath
			  to protect other folders, e.g. ~/.ssh.
			* Device shell: with --allowDeviceShell (and qual or dev mode), the agent can run single shell
			  commands on a device for troubleshooting, via "remoteaccess connect ssh --device <id> -- <command>"
			  (if remote access is configured) or c8y_Command operations. Add --confirmDeviceShell to confirm
			  each command in a native dialog, which the MCP client can't answer.
			* Approval: how write commands are approved:
			    auto    ask via MCP elicitation if the client supports it, otherwise rely on the client's tool permission prompt
			    elicit  always ask via MCP elicitation and reject writes if the client does not support it
			    client  rely on the client's tool permission prompt (c8y_write is marked as destructive)
			    none    no approval (only for trusted automation)

			Each command run by the agent is recorded in the c8y activity log with the message "mcp:<tool>".

			The policy can also be loaded from a file (json, yaml or toml):

			    mode: qual
			    approval: elicit
			    allow: [devices, inventory, alarms, events, measurements, operations, api]
			    deny: ["* delete"]
			    readCommands: ["mycompany list"]
			    denyPaths: ["~/.ssh"]
			    timeout: 2m
			    maxOutputBytes: 100000
		`),
		Example: heredoc.Doc(`
			$ claude mcp add --scope user c8y -- c8y mcp serve
			Add the server to Claude Code (run once). Select the session in the browser on the first tool call

			$ claude mcp add --scope user c8y -- c8y mcp serve --session my-dev-tenant --mode qual
			Use a specific session, and at most allow creating and updating data

			$ claude mcp add --scope user c8y -- c8y mcp serve --deny "users" --deny "tenants" --approval elicit
			Deny user and tenant management, and always ask the user to approve writes

			$ claude mcp add c8y -e C8Y_HOST=https://example.cumulocity.com -e C8Y_USER=ci -e C8Y_PASSWORD=... -- c8y mcp serve
			Use credentials from environment variables (e.g. for CI)
		`),
		Args: cobra.NoArgs,
		RunE: ccmd.RunE,
	}

	cmd.Flags().StringVar(&ccmd.policyFile, "policy", "", "Policy file (json, yaml or toml). Flags override the values in the file")
	cmd.Flags().StringVar(&ccmd.mode, "mode", "", "Maximum mode the agent can use: prod (read only), qual (create, update) or dev (create, update, delete). Default: the session's mode (prod if the session has none)")
	cmd.Flags().StringArrayVar(&ccmd.allow, "allow", nil, "Only allow commands matching this pattern, e.g. devices, \"inventory *\". Can be used multiple times")
	cmd.Flags().StringArrayVar(&ccmd.deny, "deny", nil, "Deny commands matching this pattern, e.g. users, \"* delete\". Can be used multiple times")
	cmd.Flags().StringArrayVar(&ccmd.denyPaths, "denyPath", nil, "Local file or folder the commands can't read, e.g. ~/.ssh. The session home folder and the session file are always denied. Can be used multiple times")
	cmd.Flags().StringArrayVar(&ccmd.readCommands, "read", nil, "Additional commands to treat as read-only, e.g. extension commands. Can be used multiple times")
	cmd.Flags().StringVar(&ccmd.approval, "approval", "", "How write commands are approved: auto, elicit, client or none. Default: auto")
	cmd.Flags().DurationVar(&ccmd.timeout, "timeout", 0, "Default timeout per command. Default: 2m")
	cmd.Flags().DurationVar(&ccmd.maxTimeout, "maxTimeout", 0, "Maximum timeout the agent can request per command. Default: 10m")
	cmd.Flags().IntVar(&ccmd.maxOutput, "maxOutput", 0, "Maximum number of output bytes returned to the agent per command. Default: 100000")
	// stdin/stdout is always used. The flag is accepted for compatibility with existing registrations
	cmd.Flags().BoolVar(&ccmd.stdio, "stdio", true, "Communicate over stdin/stdout (always enabled)")
	_ = cmd.Flags().MarkHidden("stdio")
	cmd.Flags().BoolVar(&ccmd.allowDeviceShell, "allowDeviceShell", false, "Allow the agent to run shell commands on devices (remoteaccess connect ssh and c8y_Command operations). Needs --mode qual or dev")
	cmd.Flags().BoolVar(&ccmd.confirmDeviceShell, "confirmDeviceShell", false, "Ask for confirmation of each command run on a device in a native dialog, in addition to the normal approval")
	cmd.Flags().BoolVar(&ccmd.noNotify, "noNotify", false, "Don't show a desktop notification when the session picker opens in the browser")
	cmd.Flags().BoolVar(&ccmd.noDefaultDeny, "noDefaultDeny", false, fmt.Sprintf("Don't include the default deny list: %s", strings.Join(mcpserver.DefaultDenyCommands, ", ")))

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("mode", "prod\tread only", "qual\tcreate and update", "dev\tcreate, update and delete"),
		completion.WithValidateSet("approval", "auto", "elicit", "client", "none"),
	)

	cmdutil.DisableEncryptionCheck(cmd)
	cmdutil.DisableAuthCheck(cmd)
	cmd.SilenceUsage = true
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

type policyFile struct {
	Mode           string        `mapstructure:"mode"`
	Allow          []string      `mapstructure:"allow"`
	Deny           []string      `mapstructure:"deny"`
	DenyPaths      []string      `mapstructure:"denyPaths"`
	ReadCommands   []string      `mapstructure:"readCommands"`
	Approval       string        `mapstructure:"approval"`
	Timeout        time.Duration `mapstructure:"timeout"`
	MaxTimeout     time.Duration `mapstructure:"maxTimeout"`
	MaxOutputBytes int           `mapstructure:"maxOutputBytes"`
	NoDefaultDeny  bool          `mapstructure:"noDefaultDeny"`

	AllowDeviceShell   bool `mapstructure:"allowDeviceShell"`
	ConfirmDeviceShell bool `mapstructure:"confirmDeviceShell"`
}

func (n *CmdServe) buildPolicy(cmd *cobra.Command) (mcpserver.Policy, error) {
	policy := mcpserver.DefaultPolicy()
	file := policyFile{}

	if n.policyFile != "" {
		v := viper.New()
		v.SetConfigFile(n.policyFile)
		if err := v.ReadInConfig(); err != nil {
			return policy, cmderrors.NewUserError(fmt.Sprintf("could not read policy file. %s", err))
		}
		if err := v.Unmarshal(&file); err != nil {
			return policy, cmderrors.NewUserError(fmt.Sprintf("invalid policy file. %s", err))
		}
	}

	pick := func(flagName string, flagValue string, fileValue string) string {
		if cmd.Flags().Changed(flagName) {
			return flagValue
		}
		return fileValue
	}

	if v := pick("mode", n.mode, file.Mode); v != "" {
		mode, err := mcpserver.ParseMode(v)
		if err != nil {
			return policy, cmderrors.NewUserError(err)
		}
		policy.Mode = mode
	}

	approval, err := mcpserver.ParseApproval(pick("approval", n.approval, file.Approval))
	if err != nil {
		return policy, cmderrors.NewUserError(err)
	}
	policy.Approval = approval

	if n.noDefaultDeny || file.NoDefaultDeny {
		policy.Deny = nil
	}
	policy.Deny = append(policy.Deny, file.Deny...)
	policy.Deny = append(policy.Deny, n.deny...)
	policy.Allow = append(append(policy.Allow, file.Allow...), n.allow...)
	for _, p := range append(append([]string{}, file.DenyPaths...), n.denyPaths...) {
		policy.DenyPaths = append(policy.DenyPaths, expandHome(p))
	}
	policy.ReadCommands = append(append(policy.ReadCommands, file.ReadCommands...), n.readCommands...)

	if cmd.Flags().Changed("timeout") {
		policy.Timeout = n.timeout
	} else if file.Timeout > 0 {
		policy.Timeout = file.Timeout
	}
	if cmd.Flags().Changed("maxTimeout") {
		policy.MaxTimeout = n.maxTimeout
	} else if file.MaxTimeout > 0 {
		policy.MaxTimeout = file.MaxTimeout
	}
	if cmd.Flags().Changed("maxOutput") {
		policy.MaxOutputBytes = n.maxOutput
	} else if file.MaxOutputBytes > 0 {
		policy.MaxOutputBytes = file.MaxOutputBytes
	}
	if policy.MaxOutputBytes <= 0 {
		return policy, cmderrors.NewUserError("maxOutput must be greater than 0")
	}
	policy.AllowDeviceShell = n.allowDeviceShell || file.AllowDeviceShell
	n.confirmDeviceShell = n.confirmDeviceShell || file.ConfirmDeviceShell
	if n.confirmDeviceShell && !policy.AllowDeviceShell {
		return policy, cmderrors.NewUserError("--confirmDeviceShell needs --allowDeviceShell")
	}
	return policy, nil
}

func (n *CmdServe) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	policy, err := n.buildPolicy(cmd)
	if err != nil {
		return err
	}

	if n.factory.IOStreams.IsStdinTTY() {
		return cmderrors.NewUserError(heredoc.Doc(`
			c8y mcp serve is started by the MCP client, which communicates with it over stdin/stdout.
			Register it in the MCP client instead of running it in a terminal, e.g. for Claude Code:

			    claude mcp add --scope user c8y -- c8y mcp serve
		`))
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find the c8y executable. %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	host := cfg.GetHost()
	sessionFile := mcpserver.ResolveSessionFile(cfg.GetSessionFile(), cfg.GetSessionHomeDir())
	// Without a session, the user selects one in the browser on the first tool call
	// Read the mode from the resolved session file, the same way as when switching sessions.
	// Sessions which are only defined by environment variables use the current configuration
	sessionModeName := cfg.GetString(config.SettingsMode)
	if info, err := os.Stat(sessionFile); err == nil && !info.IsDir() {
		sessionModeName = readSessionMode(sessionFile)
	}

	server := &mcpserver.Server{
		Version: n.factory.BuildVersion,
		Policy:  policy,
		Classifier: &mcpserver.Classifier{
			Root:   cmd.Root(),
			Policy: policy,
		},
	}

	return n.runStdio(ctx, cfg, server, executable, host, sessionFile, sessionModeName)
}

// capMode returns the session mode, and the mode the agent can use (the policy mode limited
// by the session mode). The mode stored in the session is used, ignoring CI detection, so CI
// env variables can't widen what the agent is allowed to do
func capMode(log interface{ Warnf(string, ...any) }, policyMode config.SessionMode, sessionModeName string) (config.SessionMode, config.SessionMode) {
	sessionMode := config.SessionModeUnset.FromString(sessionModeName, false)
	if sessionMode == config.SessionModeUnset {
		sessionMode = config.SessionModeProduction
	}
	effectiveMode := mcpserver.CapMode(policyMode, sessionMode)
	if policyMode != config.SessionModeUnset && effectiveMode != policyMode {
		log.Warnf("MCP mode %q is limited by the session mode %q", policyMode.String(), sessionMode.String())
	}
	return sessionMode, effectiveMode
}

// expandHome expands a leading ~ to the user's home folder
func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[1:])
		}
	}
	return p
}

// deniedPaths returns the local paths the commands can't access: the session home
// folder (with all session files and the encryption key), the session file and the
// paths from the policy
func deniedPaths(cfg *config.Config, policy mcpserver.Policy, sessionFile string) []string {
	paths := []string{}
	if home := cfg.GetSessionHomeDir(); home != "" {
		paths = append(paths, home)
	}
	if sessionFile != "" && isFile(sessionFile) {
		paths = append(paths, sessionFile)
	}
	// keep the paths the user denied in the settings, as the env variable
	// set for the commands replaces the setting
	paths = append(paths, cfg.GetDeniedPaths()...)
	return append(paths, policy.DenyPaths...)
}

func isFile(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

type sessionDetails struct {
	host, tenant, username string
	// mode the session's mode (settings.session.mode)
	mode string
	// encrypted an encrypted secret of the session (if any), to check the passphrase
	encrypted string
}

// readSessionDetails returns the non-secret details of a session file
func readSessionDetails(sessionFile string) sessionDetails {
	c := config.NewConfig(viper.New())
	c.SetSessionFile(sessionFile)
	_, _ = c.ReadConfigFiles(nil)
	details := sessionDetails{
		host:     c.GetHost(),
		tenant:   c.GetTenant(),
		username: c.GetUsername(),
		mode:     c.GetString(config.SettingsMode),
	}
	for _, key := range []string{"password", config.SettingsToken} {
		if value := c.GetString(key); c.SecureData.IsEncrypted(value) == 1 {
			details.encrypted = value
			break
		}
	}
	return details
}

// readSessionMode returns the mode of a session. A new configuration is used, so no
// settings are inherited from the previously used session
func readSessionMode(sessionFile string) string {
	c := config.NewConfig(viper.New())
	c.SetSessionFile(sessionFile)
	_, _ = c.ReadConfigFiles(nil)
	return c.GetString(config.SettingsMode)
}
