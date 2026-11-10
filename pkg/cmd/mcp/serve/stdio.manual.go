package serve

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/sessions/selectsession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mcpserver"
)

// runStdio runs the server over stdin/stdout, where the MCP client starts the server
// and there is no terminal
func (n *CmdServe) runStdio(ctx context.Context, cfg *config.Config, server *mcpserver.Server, executable string, host string, sessionFile string, sessionModeName string) error {
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	// the server's log is shown by the MCP client (e.g. claude --debug)
	server.OnCall = func(tool string, command string, outcome string) {
		log.Infof("%s: %s (%s)", tool, command, outcome)
	}
	if n.confirmDeviceShell {
		// stdin is used by the MCP transport, so only a dialog can be used
		confirmer := &mcpserver.DeviceShellConfirmer{}
		server.ConfirmDeviceShell = confirmer.Confirm
	}

	// Session only defined by environment variables (e.g. CI). The commands use
	// the server's environment
	if !isFile(sessionFile) && host != "" {
		sessionMode, effectiveMode := capMode(log, server.Policy.Mode, sessionModeName)
		server.Session = mcpserver.SessionInfo{
			Host:        host,
			Tenant:      cfg.GetTenant(),
			Username:    cfg.GetUsername(),
			SessionName: mcpserver.SessionKey(sessionFile),
			SessionMode: sessionMode.String(),
		}
		server.Runner = &mcpserver.Runner{Executable: executable, SessionFile: sessionFile, Mode: effectiveMode}
		log.Infof("Starting MCP server (stdio). host=%s, mode=%s, approval=%s", host, effectiveMode.String(), server.Policy.Approval)
		return runStdioServer(ctx, server)
	}

	tempDir, err := os.MkdirTemp("", "c8y-mcp-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	sessions := &stdioSessions{
		server:     server,
		cfg:        cfg,
		log:        log,
		executable: executable,
		tempDir:    tempDir,
		stderr:     n.factory.IOStreams.ErrOut,
		picker:     &mcpserver.WebPicker{},
	}
	if !n.noNotify {
		sessions.picker.Notify = func(message string) {
			mcpserver.DesktopNotification("c8y mcp serve", message)
		}
	}
	if isFile(sessionFile) {
		if sessions.needsPassphrase(sessionFile) {
			// ask for the passphrase in the browser on the first tool call
			sessions.startup = sessionFile
		} else {
			// log in on the first command (with the pin entry command, if needed)
			session, _, err := sessions.newSession(sessionFile)
			if err != nil {
				return err
			}
			sessions.use(session)
		}
	}
	// select a session on the first tool call (if none was given), and wait for a
	// session switch which is in progress
	server.Gate = func(ctx context.Context, ss *mcp.ServerSession) error {
		return sessions.ensure(ctx)
	}
	server.SwitchSession = sessions.startSwitch

	log.Infof("Starting MCP server (stdio). session=%s, approval=%s", sessionFile, server.Policy.Approval)
	return runStdioServer(ctx, server)
}

func runStdioServer(ctx context.Context, server *mcpserver.Server) error {
	if err := server.New().Run(ctx, &mcp.StdioTransport{}); err != nil && ctx.Err() == nil {
		return err
	}
	return nil
}

// stdioSessions selects sessions in the browser and logs in to them, as there is no
// terminal in stdio mode
type stdioSessions struct {
	server     *mcpserver.Server
	cfg        *config.Config
	log        *logger.Logger
	executable string
	tempDir    string
	stderr     io.Writer
	picker     *mcpserver.WebPicker

	// mu is held while a session is selected, so tool calls wait for the selection
	mu      sync.Mutex
	current string
	// startup session given with --session, which needs the passphrase from the browser
	startup string

	// sessions which need a two-factor code
	codeMu    sync.Mutex
	needsCode map[string]bool
}

// pendingSession a session which was logged in to, but isn't used yet
type pendingSession struct {
	info   mcpserver.SessionInfo
	runner *mcpserver.Runner
	path   string
}

// newSession prepares the login to a session file. The login happens when the
// returned runner runs its first command, unless it is done before
func (s *stdioSessions) newSession(sessionFile string) (*pendingSession, *mcpserver.InteractiveLogin, error) {
	details := readSessionDetails(sessionFile)
	if details.host == "" {
		return nil, nil, cmderrors.NewUserError(fmt.Sprintf("the session does not have a host: %s", sessionFile))
	}
	sessionMode, effectiveMode := capMode(s.log, s.server.Policy.Mode, details.mode)

	tempDir, err := os.MkdirTemp(s.tempDir, "session-")
	if err != nil {
		return nil, nil, err
	}
	login := &mcpserver.InteractiveLogin{
		Executable:  s.executable,
		SessionFile: sessionFile,
		// there is no terminal to select the mode
		SessionMode: sessionMode.String(),
		Stderr:      s.stderr,
		TempDir:     tempDir,
		// log in again when the token expires, asking for the passphrase or the
		// two-factor code in the browser if needed
		Renew: s.renew(sessionFile),
	}
	return &pendingSession{
		path: sessionFile,
		info: mcpserver.SessionInfo{
			Host:        details.host,
			Tenant:      details.tenant,
			Username:    details.username,
			SessionName: mcpserver.SessionKey(sessionFile),
			SessionMode: sessionMode.String(),
		},
		runner: &mcpserver.Runner{
			Executable:  s.executable,
			SessionFile: sessionFile,
			Mode:        effectiveMode,
			Auth:        login,
		},
	}, login, nil
}

// use makes the session the one used by new commands
func (s *stdioSessions) use(session *pendingSession) {
	s.server.SetSession(session.info, session.runner)
	s.current = session.path
	s.startup = ""
}

// login is called by the picker when the user selects a session, so the user sees
// straight away whether the session works
func (s *stdioSessions) login(ctx context.Context, choice mcpserver.SessionChoice, secrets mcpserver.LoginSecrets) (any, error) {
	if choice.NeedsPassphrase {
		// check the passphrase first, for a clear message if it is wrong
		if err := s.checkPassphrase(choice.Path, secrets.Passphrase); err != nil {
			return nil, err
		}
	}
	session, login, err := s.newSession(choice.Path)
	if err != nil {
		return nil, err
	}
	if err := s.loginWith(ctx, choice.Path, login.LoginWith, secrets); err != nil {
		return nil, err
	}
	return session, nil
}

// loginWith logs in, and remembers whether the session needs a two-factor code, so
// the code field is shown straight away the next time
func (s *stdioSessions) loginWith(ctx context.Context, sessionFile string, login func(context.Context, mcpserver.LoginSecrets) error, secrets mcpserver.LoginSecrets) error {
	err := login(ctx, secrets)
	if errors.Is(err, mcpserver.ErrTOTPRequired) || (err == nil && secrets.TOTPCode != "") {
		s.codeMu.Lock()
		if s.needsCode == nil {
			s.needsCode = map[string]bool{}
		}
		s.needsCode[sessionFile] = true
		s.codeMu.Unlock()
	}
	return err
}

func (s *stdioSessions) needsCodeFor(sessionFile string) bool {
	s.codeMu.Lock()
	defer s.codeMu.Unlock()
	return s.needsCode[sessionFile]
}

// renew returns a function which logs in again, e.g. when the token expires. It logs
// in without user input if possible (e.g. with the pin entry command), and otherwise
// asks for the passphrase or the two-factor code in the browser
func (s *stdioSessions) renew(sessionFile string) func(context.Context, func(context.Context, mcpserver.LoginSecrets) error) error {
	return func(ctx context.Context, login func(context.Context, mcpserver.LoginSecrets) error) error {
		if !s.needsPassphrase(sessionFile) && !s.needsCodeFor(sessionFile) {
			err := s.loginWith(ctx, sessionFile, login, mcpserver.LoginSecrets{})
			if !errors.Is(err, mcpserver.ErrTOTPRequired) {
				return err
			}
		}
		picker := *s.picker
		picker.Login = func(ctx context.Context, choice mcpserver.SessionChoice, secrets mcpserver.LoginSecrets) (any, error) {
			if choice.NeedsPassphrase {
				if err := s.checkPassphrase(sessionFile, secrets.Passphrase); err != nil {
					return nil, err
				}
			}
			return nil, s.loginWith(ctx, sessionFile, login, secrets)
		}
		_, err := picker.Unlock(ctx, "Log in to the Cumulocity session", "", s.choice(sessionFile))
		return err
	}
}

// needsPassphrase returns true if the passphrase of the session must be entered in
// the browser: the session is encrypted, and neither a pin entry command nor
// C8Y_PASSPHRASE is set
func (s *stdioSessions) needsPassphrase(sessionFile string) bool {
	return s.needsPassphraseFor(readSessionDetails(sessionFile))
}

func (s *stdioSessions) needsPassphraseFor(details sessionDetails) bool {
	if s.cfg.PinEntry() != "" || os.Getenv(config.EnvPassphrase) != "" {
		return false
	}
	return details.encrypted != ""
}

// checkPassphrase checks the passphrase against the session's encrypted values
func (s *stdioSessions) checkPassphrase(sessionFile string, passphrase string) error {
	value := readSessionDetails(sessionFile).encrypted
	if value == "" {
		return nil
	}
	if _, err := s.cfg.SecureData.DecryptString(value, passphrase); err != nil {
		return errors.New("the passphrase is not correct")
	}
	return nil
}

// choice returns the details of a session shown in the picker
func (s *stdioSessions) choice(sessionFile string) mcpserver.SessionChoice {
	details := readSessionDetails(sessionFile)
	mode := mcpserver.CapMode(s.server.Policy.Mode, config.SessionModeUnset.FromString(details.mode, false))
	return mcpserver.SessionChoice{
		Path:            sessionFile,
		Name:            mcpserver.SessionKey(sessionFile),
		Host:            strings.TrimRight(details.host, "/"),
		Tenant:          details.tenant,
		Username:        details.username,
		Mode:            mode.String(),
		ModeInfo:        modeLabel(mode),
		NeedsPassphrase: s.needsPassphraseFor(details),
		NeedsCode:       s.needsCodeFor(sessionFile),
	}
}

// modeLabel describes what the agent can do in a mode
func modeLabel(mode config.SessionMode) string {
	switch mode {
	case config.SessionModeQual:
		return "read, create and update"
	case config.SessionModeDev:
		return "read, create, update and delete"
	default:
		return "read only"
	}
}

func (s *stdioSessions) choices() ([]mcpserver.SessionChoice, error) {
	list, err := selectsession.ListSessions(s.cfg, s.log)
	if err != nil {
		return nil, err
	}
	current := s.current
	if resolved, err := filepath.EvalSymlinks(current); err == nil {
		current = resolved
	}
	choices := make([]mcpserver.SessionChoice, 0, len(list))
	for _, session := range list {
		path := session.Path
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			path = resolved
		}
		choice := s.choice(session.Path)
		choice.Current = current != "" && path == current
		choices = append(choices, choice)
	}
	return choices, nil
}

// pickerWithLogin returns the picker, which logs in to the selected session
func (s *stdioSessions) pickerWithLogin() *mcpserver.WebPicker {
	picker := *s.picker
	picker.Login = s.login
	return &picker
}

// pick lets the user select a session in the browser and log in. It must be called
// with mu held
func (s *stdioSessions) pick(ctx context.Context, title string) error {
	choices, err := s.choices()
	if err != nil {
		return err
	}
	selection, err := s.pickerWithLogin().Pick(ctx, title, choices)
	if err != nil {
		return err
	}
	s.use(selection.Value.(*pendingSession))
	return nil
}

// unlockStartup asks for the passphrase of the session given with --session. It must
// be called with mu held
func (s *stdioSessions) unlockStartup(ctx context.Context) error {
	selection, err := s.pickerWithLogin().Unlock(ctx, "Unlock the Cumulocity session", "", s.choice(s.startup))
	if err != nil {
		return err
	}
	s.use(selection.Value.(*pendingSession))
	return nil
}

// ensure makes sure a session is selected, waiting for a selection in progress
func (s *stdioSessions) ensure(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != "" {
		return nil
	}
	var err error
	if s.startup != "" {
		err = s.unlockStartup(ctx)
	} else {
		err = s.pick(ctx, "Select a Cumulocity session")
	}
	if err != nil {
		if errors.Is(err, mcpserver.ErrPickerCancelled) || errors.Is(err, mcpserver.ErrPickerTimeout) {
			return fmt.Errorf("no Cumulocity session was selected (%w). Ask the user which session to use; the session picker opens in the browser on the next c8y tool call", err)
		}
		return fmt.Errorf("could not select a Cumulocity session. %w. Ask the user to configure the session in the MCP server (--session) instead", err)
	}
	return nil
}

// startSwitch opens the session picker in the background. Tool calls wait until the
// user has selected a session (or cancelled)
func (s *stdioSessions) startSwitch(ctx context.Context) (string, error) {
	if !s.mu.TryLock() {
		return "A Cumulocity session picker is already open in the browser. Wait for me to select a session.", nil
	}
	go func() {
		defer s.mu.Unlock()
		if err := s.pick(context.Background(), "Switch the Cumulocity session"); err != nil && !errors.Is(err, mcpserver.ErrPickerCancelled) {
			s.log.Warnf("Session was not switched. %s", err)
		}
	}()
	return fmt.Sprintf("I'm switching the Cumulocity session used by the %s tools in the browser window which just opened. "+
		"Once I've selected it, call %s and tell me which tenant and user are now used. Don't change any data as part of this.",
		"c8y", mcpserver.ToolSession), nil
}
