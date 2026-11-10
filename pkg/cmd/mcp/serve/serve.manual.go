package serve

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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
	readCommands  []string
	approval      string
	timeout       time.Duration
	maxTimeout    time.Duration
	maxOutput     int
	noDefaultDeny bool
	noNotify      bool

	allowDeviceShell   bool
	confirmDeviceShell bool
	deviceShellInfo    string
	stdio              bool
	socket             string
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

			By default you run the server in your own terminal. It logs in using the normal session login,
			so encrypted sessions, two-factor authentication and SSO work as usual, then listens on a
			unix socket which only your user can access. If no session is set (via --session or C8Y_SESSION),
			you are prompted to select one. MCP clients connect to it with "c8y mcp connect".
			The token is renewed by logging in again in the terminal when it is about to expire.

			You can switch to another session while the server is running by typing "s" and Enter in the
			server terminal. Connected clients use the new session for new commands, so the MCP client only
			needs to be registered once. Type "i" to show the current session and "h" for help.

			Each new client must be approved in the server terminal, which shows the process that is connecting
			and the processes that started it. This stops other programs running as your user from using the
			server without you noticing, as the approval can't be given by the client. The client connects
			straight away, and its tool calls wait until it is approved, so it doesn't time out. The terminal
			bell and a desktop notification draw your attention to the prompt (disable with --noNotify).
			One answer applies to all connections from the same client which are waiting, and answering "a"
			also allows later connections from that client (e.g. reconnects), until it hasn't connected for
			8 hours. "a" is only offered when the connecting process is this c8y binary (c8y mcp connect).
			The socket must be in a directory which is owned by your user and not writable by other users.

			With --stdio, the MCP client starts the server itself and communicates over stdin/stdout. The
			session must then work without user input, e.g. by setting C8Y_HOST, C8Y_TENANT, C8Y_USER and
			C8Y_PASSWORD or C8Y_TOKEN in the MCP client configuration, or C8Y_PASSPHRASE for encrypted
			sessions. This mode is intended for automation and CI.

			You control what the agent can do:

			* Session: selected by you in the server terminal. The agent can't switch the session, host or
			  credentials. In socket mode, commands only receive the token, not the password or passphrase.
			* Mode (prod, qual, dev): the maximum mode the agent can use. It can only narrow the session's own
			  mode, never widen it, also after switching the session. Read-only commands always run in prod
			  (read-only) mode.
			* Allow and deny lists: which commands the agent can use, e.g. "devices", "inventory *", "* delete".
			* Device shell: with --allowDeviceShell (and qual or dev mode), the agent can run single shell
			  commands on a device for troubleshooting, via "remoteaccess connect ssh --device <id> -- <command>"
			  (if remote access is configured) or c8y_Command operations. Add --confirmDeviceShell to confirm
			  each command in a dialog or the server terminal, which the MCP client can't answer.
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
			    timeout: 2m
			    maxOutputBytes: 100000
		`),
		Example: heredoc.Doc(`
			$ c8y mcp serve
			Select a session interactively, log in and start the server in the current terminal (read only)

			$ c8y mcp serve --session my-dev-tenant --mode qual
			Log in to the session and start the server in the current terminal (allowing create and update commands)

			$ claude mcp add --scope user c8y -- c8y mcp connect
			Add the server to Claude Code (run once). Claude Code starts the relay, which connects to the running server

			$ c8y mcp serve --session dev --mode dev --deny "users" --deny "tenants" --approval elicit
			Allow all actions (including delete) except for user and tenant management, and always ask the user to approve writes

			$ claude mcp add c8y -e C8Y_HOST=https://example.cumulocity.com -e C8Y_USER=ci -e C8Y_PASSWORD=... -- c8y mcp serve --stdio
			Let the MCP client start the server, using credentials from environment variables
		`),
		Args: cobra.NoArgs,
		RunE: ccmd.RunE,
	}

	cmd.Flags().StringVar(&ccmd.policyFile, "policy", "", "Policy file (json, yaml or toml). Flags override the values in the file")
	cmd.Flags().StringVar(&ccmd.mode, "mode", "", "Maximum mode the agent can use: prod (read only), qual (create, update) or dev (create, update, delete). Default: prod")
	cmd.Flags().StringArrayVar(&ccmd.allow, "allow", nil, "Only allow commands matching this pattern, e.g. devices, \"inventory *\". Can be used multiple times")
	cmd.Flags().StringArrayVar(&ccmd.deny, "deny", nil, "Deny commands matching this pattern, e.g. users, \"* delete\". Can be used multiple times")
	cmd.Flags().StringArrayVar(&ccmd.readCommands, "read", nil, "Additional commands to treat as read-only, e.g. extension commands. Can be used multiple times")
	cmd.Flags().StringVar(&ccmd.approval, "approval", "", "How write commands are approved: auto, elicit, client or none. Default: auto")
	cmd.Flags().DurationVar(&ccmd.timeout, "timeout", 0, "Default timeout per command. Default: 2m")
	cmd.Flags().DurationVar(&ccmd.maxTimeout, "maxTimeout", 0, "Maximum timeout the agent can request per command. Default: 10m")
	cmd.Flags().IntVar(&ccmd.maxOutput, "maxOutput", 0, "Maximum number of output bytes returned to the agent per command. Default: 100000")
	cmd.Flags().BoolVar(&ccmd.stdio, "stdio", false, "Communicate over stdin/stdout (the MCP client starts the server). The session must work without user input")
	cmd.Flags().StringVar(&ccmd.socket, "socket", "", "Unix socket path to listen on, e.g. to run several servers at the same time. Defaults to server.sock in the mcp folder of the session home directory")
	cmd.Flags().BoolVar(&ccmd.allowDeviceShell, "allowDeviceShell", false, "Allow the agent to run shell commands on devices (remoteaccess connect ssh and c8y_Command operations). Needs --mode qual or dev")
	cmd.Flags().BoolVar(&ccmd.confirmDeviceShell, "confirmDeviceShell", false, "Ask for confirmation of each command run on a device in a dialog or the server terminal, in addition to the normal approval")
	cmd.Flags().BoolVar(&ccmd.noNotify, "noNotify", false, "Don't show a desktop notification when a client needs to be approved")
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
	switch {
	case n.confirmDeviceShell:
		n.deviceShellInfo = "shell commands allowed, each one is confirmed by you here or in a dialog"
	case policy.AllowDeviceShell:
		n.deviceShellInfo = "shell commands allowed (approved like other write commands)"
	default:
		n.deviceShellInfo = "shell commands disabled (enable with --allowDeviceShell)"
	}
	return policy, nil
}

func (n *CmdServe) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}

	policy, err := n.buildPolicy(cmd)
	if err != nil {
		return err
	}

	if !n.stdio && !n.factory.IOStreams.IsStdinTTY() {
		return cmderrors.NewUserError(heredoc.Doc(`
			c8y mcp serve needs a terminal to log in. Run it in a terminal, and register "c8y mcp connect" in the
			MCP client instead. Use --stdio if the MCP client should start the server itself.
		`))
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find the c8y executable. %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Let the user pick a session interactively if none is configured (similar to set-session).
	// This is only possible in socket mode, as stdin is used by the MCP transport in stdio mode
	if !n.stdio && cfg.GetHost() == "" && cfg.GetSessionFile() == "" {
		selected, err := mcpserver.SelectSession(ctx, executable, os.Stdin, n.factory.IOStreams.ErrOut)
		if err != nil {
			return cmderrors.NewUserError(err)
		}
		if selected == "" {
			return cmderrors.NewUserError("no Cumulocity session was selected")
		}
		cfg.SetSessionFile(selected)
		if _, err := cfg.ReadConfigFiles(nil); err != nil {
			return err
		}
	}

	host := cfg.GetHost()
	sessionFile := mcpserver.ResolveSessionFile(cfg.GetSessionFile(), cfg.GetSessionHomeDir())
	if host == "" && sessionFile == "" {
		return cmderrors.NewUserError("no Cumulocity session is configured. Use --session <name> or set C8Y_SESSION")
	}
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

	if n.stdio {
		if host == "" {
			return cmderrors.NewUserError("the session does not have a host")
		}
		sessionMode, effectiveMode := capMode(log, policy.Mode, sessionModeName)
		server.Session = mcpserver.SessionInfo{
			Host:        host,
			Tenant:      cfg.GetTenant(),
			Username:    cfg.GetUsername(),
			SessionName: mcpserver.SessionKey(sessionFile),
			SessionMode: sessionMode.String(),
		}
		server.Runner = &mcpserver.Runner{
			Executable:  executable,
			SessionFile: sessionFile,
			Mode:        effectiveMode,
		}
		if n.confirmDeviceShell {
			// stdin is used by the MCP transport, so only a dialog can be used
			confirmer := &mcpserver.DeviceShellConfirmer{Stderr: n.factory.IOStreams.ErrOut}
			server.ConfirmDeviceShell = confirmer.Confirm
		}
		log.Infof("Starting MCP server (stdio). host=%s, mode=%s, approval=%s", host, effectiveMode.String(), policy.Approval)
		if err := server.New().Run(ctx, &mcp.StdioTransport{}); err != nil && ctx.Err() == nil {
			return err
		}
		return nil
	}
	return n.runSocket(ctx, cfg, server, executable, sessionFile, sessionModeName)
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
	if effectiveMode != policyMode {
		log.Warnf("MCP mode %q is limited by the session mode %q", policyMode.String(), sessionMode.String())
	}
	return sessionMode, effectiveMode
}

// readSessionMode returns the mode of a session. A new configuration is used, so no
// settings are inherited from the previously used session
func readSessionMode(sessionFile string) string {
	c := config.NewConfig(viper.New())
	c.SetSessionFile(sessionFile)
	_, _ = c.ReadConfigFiles(nil)
	return c.GetString(config.SettingsMode)
}

// activeSession a session which the server has logged in to
type activeSession struct {
	info   mcpserver.SessionInfo
	runner *mcpserver.Runner
	login  *mcpserver.InteractiveLogin
}

// sessionStarter logs in to sessions in the server's terminal
type sessionStarter struct {
	executable string
	policyMode config.SessionMode
	tempDir    string
	terminal   sync.Locker
	stderr     io.Writer
	log        interface{ Warnf(string, ...any) }
}

// start logs in to the session (prompting in the terminal if needed) and returns the
// session details and the runner which uses it
func (s *sessionStarter) start(ctx context.Context, sessionFile string, sessionModeName string) (*activeSession, error) {
	// Each session has its own directory for the copy of its session file
	tempDir, err := os.MkdirTemp(s.tempDir, "session-")
	if err != nil {
		return nil, err
	}

	login := &mcpserver.InteractiveLogin{
		Executable:  s.executable,
		SessionFile: sessionFile,
		SessionMode: sessionModeName,
		Stdin:       os.Stdin,
		Stderr:      s.stderr,
		TempDir:     tempDir,
		Terminal:    s.terminal,
	}
	if err := login.Login(ctx); err != nil {
		os.RemoveAll(tempDir)
		return nil, err
	}

	sessionMode, effectiveMode := capMode(s.log, s.policyMode, sessionModeName)
	info := mcpserver.SessionInfo{
		SessionName: mcpserver.SessionKey(sessionFile),
		SessionMode: sessionMode.String(),
	}
	// Use the values from the login, which take environment variables into account
	if authEnv, err := login.AuthEnv(ctx); err == nil {
		for _, item := range authEnv {
			key, value, _ := strings.Cut(item, "=")
			switch key {
			case "C8Y_HOST":
				info.Host = value
			case "C8Y_TENANT":
				info.Tenant = value
			case "C8Y_USER":
				info.Username = value
			}
		}
	}

	return &activeSession{
		info:  info,
		login: login,
		runner: &mcpserver.Runner{
			Executable:  s.executable,
			SessionFile: sessionFile,
			Mode:        effectiveMode,
			Auth:        login,
		},
	}, nil
}

func (n *CmdServe) printSession(session *activeSession, approval mcpserver.Approval) {
	stderr := n.factory.IOStreams.ErrOut
	auth := "basic auth (commands receive the password)"
	if session.login.UsesToken() {
		auth = "token (expiry unknown)"
		if t := session.login.ExpiresAt(); t != nil {
			auth = "token, expires " + t.Local().Format(time.RFC1123)
		}
	}
	if session.info.SessionName != "" && session.info.SessionName != "default" {
		fmt.Fprintf(stderr, "  session:   %s\n", session.info.SessionName)
	}
	fmt.Fprintf(stderr, "  tenant:    %s (%s)\n", session.info.Host, session.info.Tenant)
	fmt.Fprintf(stderr, "  user:      %s\n", session.info.Username)
	fmt.Fprintf(stderr, "  mode:      %s\n", session.runner.Mode.Description())
	fmt.Fprintf(stderr, "  approval:  %s\n", approval)
	if n.deviceShellInfo != "" {
		fmt.Fprintf(stderr, "  devices:   %s\n", n.deviceShellInfo)
	}
	fmt.Fprintf(stderr, "  auth:      %s\n", auth)
}

const consoleHelp = `Commands (type the letter and press Enter):
  s  switch to another session
  i  show the current session
  h  show this help
Press Ctrl+C to stop the server.
`

func (n *CmdServe) runSocket(ctx context.Context, cfg *config.Config, server *mcpserver.Server, executable string, sessionFile string, sessionModeName string) error {
	stderr := n.factory.IOStreams.ErrOut
	cs := n.factory.IOStreams.ColorScheme()
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}

	socketPath := n.socket
	if socketPath == "" {
		socketPath = mcpserver.DefaultSocketPath(cfg.GetSessionHomeDir())
	}

	// Create the socket first so a second server fails before asking the user to log in
	listener, err := mcpserver.Listen(socketPath)
	if err != nil {
		return cmderrors.NewUserError(err)
	}
	defer os.Remove(socketPath)
	defer listener.Close()

	tempDir, err := os.MkdirTemp("", "c8y-mcp-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// The terminal is shared by the login, the session selection, the connection approval
	// and the commands typed by the user
	terminal := &sync.Mutex{}

	starter := &sessionStarter{
		executable: executable,
		policyMode: server.Policy.Mode,
		tempDir:    tempDir,
		terminal:   terminal,
		stderr:     stderr,
		log:        log,
	}
	session, err := starter.start(ctx, sessionFile, sessionModeName)
	if err != nil {
		return err
	}
	server.SetSession(session.info, session.runner)
	var sessionMu sync.Mutex

	server.OnCall = func(tool, command, outcome string) {
		icon := cs.SuccessIcon()
		if outcome != "ok" {
			icon = cs.WarningIcon()
		}
		fmt.Fprintf(stderr, "%s %s  [%s] %s: %s\n", time.Now().Format("15:04:05"), icon, tool, command, outcome)
	}

	connectArgs := "c8y mcp connect"
	if n.socket != "" {
		connectArgs += " --socket " + socketPath
	}

	fmt.Fprintf(stderr, "\n%s MCP server is running\n\n", cs.SuccessIcon())
	n.printSession(session, server.Policy.Approval)
	fmt.Fprintf(stderr, "  socket:    %s\n\n", socketPath)
	fmt.Fprintf(stderr, "Register it in your MCP client (once), e.g. for Claude Code:\n\n")
	fmt.Fprintf(stderr, "  claude mcp add --scope user c8y -- %s\n\n", connectArgs)
	fmt.Fprintf(stderr, "Each new client must be approved in this terminal. Until then, the agent is asked to remind you.\n")
	fmt.Fprint(stderr, consoleHelp)
	fmt.Fprintln(stderr)

	switchSession := func() {
		fmt.Fprintf(stderr, "\nSelect the session to switch to (Ctrl+C to cancel):\n")
		terminal.Lock()
		selected, err := mcpserver.SelectSession(ctx, executable, os.Stdin, stderr)
		terminal.Unlock()
		if err != nil || selected == "" {
			fmt.Fprintf(stderr, "%s Session not changed\n", cs.WarningIcon())
			return
		}
		newFile := mcpserver.ResolveSessionFile(selected, cfg.GetSessionHomeDir())
		newSession, err := starter.start(ctx, newFile, readSessionMode(newFile))
		if err != nil {
			fmt.Fprintf(stderr, "%s Could not log in to the session, so the session was not changed. %s\n", cs.FailureIcon(), err)
			return
		}
		sessionMu.Lock()
		session = newSession
		sessionMu.Unlock()
		server.SetSession(newSession.info, newSession.runner)
		fmt.Fprintf(stderr, "\n%s Switched session. New commands from the agent use:\n\n", cs.SuccessIcon())
		n.printSession(newSession, server.Policy.Approval)
		fmt.Fprintln(stderr)
	}

	go func() {
		err := mcpserver.RunConsole(ctx, os.Stdin, terminal, func(command string) {
			switch strings.ToLower(command) {
			case "":
			case "s", "switch":
				switchSession()
			case "i", "info":
				sessionMu.Lock()
				current := session
				sessionMu.Unlock()
				fmt.Fprintln(stderr)
				n.printSession(current, server.Policy.Approval)
				fmt.Fprintln(stderr)
			default:
				fmt.Fprint(stderr, consoleHelp)
			}
		})
		if err != nil && ctx.Err() == nil {
			fmt.Fprintf(stderr, "%s Commands typed in this terminal are not supported (%s). Restart the server to switch the session.\n", cs.WarningIcon(), err)
		}
	}()

	approver := &mcpserver.ConnectionApprover{
		Stdin:           os.Stdin,
		Stderr:          stderr,
		Terminal:        terminal,
		RelayExecutable: executable,
	}
	if !n.noNotify {
		approver.Notify = func(message string) {
			mcpserver.DesktopNotification("c8y mcp serve", message)
		}
	}

	if n.confirmDeviceShell {
		confirmer := &mcpserver.DeviceShellConfirmer{
			Stdin:    os.Stdin,
			Stderr:   stderr,
			Terminal: terminal,
		}
		if !n.noNotify {
			confirmer.Notify = func(message string) {
				mcpserver.DesktopNotification("c8y mcp serve", message)
			}
		}
		server.ConfirmDeviceShell = confirmer.Confirm
	}

	// Clients connect straight away, and their tool calls wait until they are approved
	gates := &mcpserver.ConnectionGates{}
	server.Gate = gates.Check

	err = mcpserver.Serve(ctx, server.New(), listener, mcpserver.ServeOptions{
		Authorize: approver.Approve,
		Gates:     gates,
		OnConnect: func(active int) {
			fmt.Fprintf(stderr, "%s %s  client connected (active: %d)\n", time.Now().Format("15:04:05"), cs.SuccessIcon(), active)
		},
		OnDisconnect: func(active int) {
			fmt.Fprintf(stderr, "%s %s  client disconnected (active: %d)\n", time.Now().Format("15:04:05"), cs.WarningIcon(), active)
		},
		OnRejected: func() {
			fmt.Fprintf(stderr, "%s %s  client rejected\n", time.Now().Format("15:04:05"), cs.WarningIcon())
		},
	})
	fmt.Fprintf(stderr, "\nMCP server stopped\n")
	return err
}
