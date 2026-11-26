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
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/sessions/login"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/sessions/selectsession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mcpserver"
)

// runStdio runs the server over stdin/stdout, where the MCP client starts the server
// and there is no terminal
func (n *CmdServe) runStdio(ctx context.Context, cfg *config.Config, server *mcpserver.Server, executable string, host string, sessionFile string, sessionModeName string, sources []providerSource) error {
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	// the server's log is shown by the MCP client (e.g. claude --debug)
	server.OnCall = func(tool string, command string, outcome string) {
		log.Infof("%s: %s (%s)", tool, command, outcome)
	}
	server.Logf = log.Infof
	server.EnableModePrompt = true
	if n.confirmDeviceShell {
		// stdin is used by the MCP transport, so only a dialog can be used
		confirmer := &mcpserver.DeviceShellConfirmer{}
		server.ConfirmDeviceShell = confirmer.Confirm
	}

	// Session only defined by environment variables (e.g. CI). The commands use
	// the server's environment
	if !isFile(sessionFile) && host != "" {
		sessionMode, effectiveMode := capMode(server, sessionFile, sessionModeName)
		info := mcpserver.SessionInfo{
			Host:        host,
			Tenant:      cfg.GetTenant(),
			Username:    cfg.GetUsername(),
			SessionName: mcpserver.SessionKey(sessionFile),
			SessionMode: sessionMode.String(),
		}
		server.SetSessionKey(info.SessionName, info, &mcpserver.Runner{Executable: executable, SessionFile: sessionFile, Mode: effectiveMode, DeniedPaths: deniedPaths(cfg, server.Policy, sessionFile), SSHFallbackUser: sshFallbackUser(cfg)})
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
		sources:    sources,
		listProviderSessions: func(ctx context.Context, ref string, args []string) (string, []extensions.ProviderSession, error) {
			provider, sessions, err := login.ListExtensionSessions(ctx, n.factory, ref, args, cfg.SessionProviderSecrets())
			name := ref
			if provider != nil {
				name = provider.Provider.Name
			}
			return name, sessions, err
		},
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
	server.SwitchSessionTo = sessions.switchTo
	server.SessionNames = sessions.sessionNames

	log.Infof("Starting MCP server (stdio). session=%s, sessionProviders=%v, approval=%s", sessionFile, providerRefs(sources), server.Policy.Approval)
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

	// sources of the sessions shown in the picker, in the order they are shown: session
	// files (see FileSessionSource) and extension session providers
	sources []providerSource
	// listProviderSessions lists the sessions of a provider, and returns the provider's name
	listProviderSessions func(ctx context.Context, ref string, args []string) (string, []extensions.ProviderSession, error)

	// mu is held while a session is selected, so tool calls wait for the selection
	mu sync.Mutex
	// current key of the session which is used (see choiceKey)
	current string
	// listed sessions of the providers, by key (see choiceKey)
	listedMu sync.Mutex
	listed   map[string]extensions.ProviderSession
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
	// key identifies the session (see choiceKey)
	key string
}

// FileSessionSource reserved session provider name of the session files in the session home folder
const FileSessionSource = extensions.FileSessionProvider

// providerSource a source of the sessions shown in the picker: the session files (Ref is
// FileSessionSource), or an extension session provider and the arguments passed to its list
// command
type providerSource struct {
	Ref  string
	Args []string
}

// IsFiles returns true if the source is the session files
func (p providerSource) IsFiles() bool {
	return p.Ref == FileSessionSource
}

func providerRefs(providers []providerSource) []string {
	refs := make([]string, 0, len(providers))
	for _, p := range providers {
		refs = append(refs, p.Ref)
	}
	return refs
}

// choiceKey identifies a session: the path of a session file, or the provider and the
// sessionUri of a provider session
func choiceKey(choice mcpserver.SessionChoice) string {
	if choice.Provider != "" {
		return providerKey(choice.Provider, choice.URI)
	}
	return choice.Path
}

func providerKey(provider string, uri string) string {
	return "provider:" + provider + "\x00" + uri
}

// newSession prepares the login to a session file. The login happens when the
// returned runner runs its first command, unless it is done before
func (s *stdioSessions) newSession(sessionFile string) (*pendingSession, *mcpserver.InteractiveLogin, error) {
	details := readSessionDetails(sessionFile)
	if details.host == "" {
		return nil, nil, cmderrors.NewUserError(fmt.Sprintf("the session does not have a host: %s", sessionFile))
	}
	sessionMode, effectiveMode := capMode(s.server, sessionFile, details.mode)

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
		Renew: s.renew(sessionFile, func() mcpserver.SessionChoice { return s.choice(sessionFile) }),
	}
	return &pendingSession{
		key: sessionFile,
		info: mcpserver.SessionInfo{
			Host:        details.host,
			Tenant:      details.tenant,
			Username:    details.username,
			SessionName: mcpserver.SessionKey(sessionFile),
			SessionMode: sessionMode.String(),
		},
		runner: &mcpserver.Runner{
			Executable:      s.executable,
			SessionFile:     sessionFile,
			Mode:            effectiveMode,
			Auth:            login,
			DeniedPaths:     deniedPaths(s.cfg, s.server.Policy, sessionFile),
			SSHFallbackUser: sshFallbackUser(s.cfg),
		},
	}, login, nil
}

// newProviderSession prepares the login to a session of an extension session provider. The
// provider returns the session (with its secrets) without user interaction, also when the
// token is renewed
func (s *stdioSessions) newProviderSession(choice mcpserver.SessionChoice) (*pendingSession, *mcpserver.InteractiveLogin, error) {
	key := choiceKey(choice)
	s.listedMu.Lock()
	listed, ok := s.listed[key]
	s.listedMu.Unlock()
	if !ok {
		return nil, nil, fmt.Errorf("unknown session %s of session provider %s", choice.URI, choice.Provider)
	}
	sessionMode, effectiveMode := capMode(s.server, key, listed.Mode)

	login := &mcpserver.InteractiveLogin{
		Executable:  s.executable,
		Provider:    choice.Provider,
		SessionURI:  choice.URI,
		SessionMode: sessionMode.String(),
		Stderr:      s.stderr,
		Renew: s.renew(key, func() mcpserver.SessionChoice {
			c := choice
			c.Current = false
			c.NeedsCode = s.needsCodeFor(key)
			return c
		}),
	}
	return &pendingSession{
		key: key,
		info: mcpserver.SessionInfo{
			Host:        listed.Host,
			Tenant:      listed.Tenant,
			Username:    listed.Username,
			SessionName: listed.DisplayName(),
			SessionMode: sessionMode.String(),
			Provider:    choice.Provider,
		},
		runner: &mcpserver.Runner{
			Executable:      s.executable,
			Mode:            effectiveMode,
			Auth:            login,
			DeniedPaths:     deniedPaths(s.cfg, s.server.Policy, ""),
			SSHFallbackUser: sshFallbackUser(s.cfg),
		},
	}, login, nil
}

// use makes the session the one used by new commands
func (s *stdioSessions) use(session *pendingSession) {
	s.server.SetSessionKey(session.key, session.info, session.runner)
	s.current = session.key
	s.startup = ""
}

// login is called by the picker when the user selects a session, so the user sees
// straight away whether the session works
func (s *stdioSessions) login(ctx context.Context, choice mcpserver.SessionChoice, secrets mcpserver.LoginSecrets) (any, error) {
	if choice.Provider != "" {
		session, login, err := s.newProviderSession(choice)
		if err != nil {
			return nil, err
		}
		if err := s.loginWith(ctx, session.key, login.LoginWith, secrets); err != nil {
			return nil, err
		}
		return session, nil
	}
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
func (s *stdioSessions) loginWith(ctx context.Context, key string, login func(context.Context, mcpserver.LoginSecrets) error, secrets mcpserver.LoginSecrets) error {
	err := login(ctx, secrets)
	if errors.Is(err, mcpserver.ErrTOTPRequired) || (err == nil && secrets.TOTPCode != "") {
		s.codeMu.Lock()
		if s.needsCode == nil {
			s.needsCode = map[string]bool{}
		}
		s.needsCode[key] = true
		s.codeMu.Unlock()
	}
	return err
}

func (s *stdioSessions) needsCodeFor(key string) bool {
	s.codeMu.Lock()
	defer s.codeMu.Unlock()
	return s.needsCode[key]
}

// renew returns a function which logs in again, e.g. when the token expires. It logs
// in without user input if possible (e.g. with the pin entry command, or a session
// provider), and otherwise asks for the passphrase or the two-factor code in the browser.
// key identifies the session, and choice returns its current details
func (s *stdioSessions) renew(key string, choice func() mcpserver.SessionChoice) func(context.Context, func(context.Context, mcpserver.LoginSecrets) error) error {
	return func(ctx context.Context, login func(context.Context, mcpserver.LoginSecrets) error) error {
		if !choice().NeedsPassphrase && !s.needsCodeFor(key) {
			err := s.loginWith(ctx, key, login, mcpserver.LoginSecrets{})
			if !errors.Is(err, mcpserver.ErrTOTPRequired) {
				return err
			}
		}
		picker := *s.picker
		picker.Login = func(ctx context.Context, choice mcpserver.SessionChoice, secrets mcpserver.LoginSecrets) (any, error) {
			if choice.NeedsPassphrase {
				if err := s.checkPassphrase(choice.Path, secrets.Passphrase); err != nil {
					return nil, err
				}
			}
			return nil, s.loginWith(ctx, key, login, secrets)
		}
		_, err := picker.Unlock(ctx, "Log in to the Cumulocity session", "", choice())
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
	mode := s.server.EffectiveMode(sessionFile, details.mode)
	return mcpserver.SessionChoice{
		Path:            sessionFile,
		Name:            mcpserver.SessionKey(sessionFile),
		Host:            strings.TrimRight(details.host, "/"),
		Tenant:          details.tenant,
		Username:        details.username,
		Mode:            mode.String(),
		ModeInfo:        modeLabel(mode),
		SessionMode:     sessionModeName(details.mode),
		MaxMode:         s.server.Policy.Ceiling().String(),
		NeedsPassphrase: s.needsPassphraseFor(details),
		NeedsCode:       s.needsCodeFor(sessionFile),
	}
}

// sessionModeName returns the session's own mode (prod if it has none)
func sessionModeName(name string) string {
	return mcpserver.SessionDefaultMode(config.SessionModeUnset.FromString(name, false)).String()
}

// modeLabel describes what the agent can do in a mode
func modeLabel(mode config.SessionMode) string {
	return mcpserver.ModeInfo(mode)
}

// choices returns the sessions shown in the picker, in the order of the sources: the session
// files and the sessions of the session providers. Sources which can't list their sessions are
// returned as problems
func (s *stdioSessions) choices(ctx context.Context) ([]mcpserver.SessionChoice, []mcpserver.SourceError, error) {
	listed := s.providerChoices(ctx)
	choices := []mcpserver.SessionChoice{}
	problems := []mcpserver.SourceError{}
	for i, source := range s.sources {
		if !source.IsFiles() {
			choices = append(choices, listed[i].choices...)
			problems = append(problems, listed[i].problems...)
			continue
		}
		fileChoices, err := s.fileChoices()
		if err != nil {
			if len(s.sources) == 1 {
				return nil, nil, err
			}
			problems = append(problems, mcpserver.SourceError{Source: mcpserver.SessionFilesSource, Message: err.Error()})
			continue
		}
		choices = append(choices, fileChoices...)
	}
	return choices, problems, nil
}

// fileChoices returns the sessions of the session files in the session home folder
func (s *stdioSessions) fileChoices() ([]mcpserver.SessionChoice, error) {
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

// providerListTimeout maximum duration to list the sessions of a session provider
const providerListTimeout = 20 * time.Second

// sourceChoices the sessions of a source, or the reason why they could not be listed
type sourceChoices struct {
	choices  []mcpserver.SessionChoice
	problems []mcpserver.SourceError
}

// providerChoices lists the sessions of the session providers (in parallel). The result has
// one item per source (empty for the session files)
func (s *stdioSessions) providerChoices(ctx context.Context) []sourceChoices {
	type result struct {
		name     string
		sessions []extensions.ProviderSession
		err      error
	}
	results := make([]result, len(s.sources))
	var wg sync.WaitGroup
	for i, provider := range s.sources {
		if provider.IsFiles() {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			listCtx, cancel := context.WithTimeout(ctx, providerListTimeout)
			defer cancel()
			name, sessions, err := s.listProviderSessions(listCtx, provider.Ref, provider.Args)
			results[i] = result{name: name, sessions: sessions, err: err}
		}()
	}
	wg.Wait()

	out := make([]sourceChoices, len(s.sources))
	listed := map[string]extensions.ProviderSession{}
	for i, r := range results {
		provider := s.sources[i]
		if provider.IsFiles() {
			continue
		}
		source := r.name
		if source == "" {
			source = provider.Ref
		}
		if r.err != nil {
			s.log.Warnf("Could not list the sessions of session provider %s. %s", provider.Ref, r.err)
			out[i].problems = append(out[i].problems, mcpserver.SourceError{Source: source, Message: r.err.Error()})
			continue
		}
		if len(r.sessions) == 0 {
			out[i].problems = append(out[i].problems, mcpserver.SourceError{Source: source, Message: "The session provider did not return any sessions."})
			continue
		}
		for _, session := range r.sessions {
			key := providerKey(provider.Ref, session.SessionURI)
			listed[key] = session
			mode := s.server.EffectiveMode(key, session.Mode)
			out[i].choices = append(out[i].choices, mcpserver.SessionChoice{
				Name:        session.DisplayName(),
				Host:        strings.TrimRight(session.Host, "/"),
				Tenant:      session.Tenant,
				Username:    session.Username,
				Mode:        mode.String(),
				ModeInfo:    modeLabel(mode),
				SessionMode: sessionModeName(session.Mode),
				MaxMode:     s.server.Policy.Ceiling().String(),
				NeedsCode:   s.needsCodeFor(key),
				Current:     s.current == key,
				Source:      source,
				Provider:    provider.Ref,
				URI:         session.SessionURI,
				Group:       session.Group,
			})
		}
	}
	s.listedMu.Lock()
	if s.listed == nil {
		s.listed = map[string]extensions.ProviderSession{}
	}
	// keep the sessions which were listed before, e.g. the current session
	for key, session := range listed {
		s.listed[key] = session
	}
	s.listedMu.Unlock()
	return out
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
	choices, problems, err := s.choices(ctx)
	if err != nil {
		return err
	}
	selection, err := s.pickerWithLogin().Pick(ctx, title, choices, problems...)
	if err != nil {
		return err
	}
	if selection.ModeOnly {
		return s.applyPickedMode(selection)
	}
	s.use(selection.Value.(*pendingSession))
	return s.applyPickedMode(selection)
}

// applyPickedMode applies the mode the user chose in the picker to the active session. The
// session's own mode resets the override, any other mode is an override up to the ceiling
func (s *stdioSessions) applyPickedMode(selection *mcpserver.Selection) error {
	if selection.Mode == config.SessionModeUnset {
		return nil
	}
	mode := selection.Mode
	if sessionModeName(selection.Choice.SessionMode) == mode.String() {
		mode = config.SessionModeUnset
	}
	_, err := s.server.SetMode(mode, "picker")
	return err
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

// sessionNames returns the names of the sessions to complete the arguments of a prompt:
// the session files, and the sessions of the providers which were listed before (the
// providers are not listed again for each completion)
func (s *stdioSessions) sessionNames(ctx context.Context) []string {
	var names []string
	seen := map[string]bool{}
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	for _, source := range s.sources {
		if source.IsFiles() {
			if list, err := selectsession.ListSessions(s.cfg, s.log); err == nil {
				for _, session := range list {
					add(mcpserver.SessionKey(session.Path))
				}
			}
		}
	}
	s.listedMu.Lock()
	for _, session := range s.listed {
		add(session.DisplayName())
	}
	s.listedMu.Unlock()
	return names
}

// findChoice returns the session with the given name (or session file path, or provider
// session uri), and fails if the name is unknown or ambiguous
func (s *stdioSessions) findChoice(ctx context.Context, name string) (mcpserver.SessionChoice, error) {
	matches := func(choice mcpserver.SessionChoice) bool {
		return strings.EqualFold(choice.Name, name) || (choice.Path != "" && choice.Path == name) || (choice.URI != "" && choice.URI == name)
	}
	var found []mcpserver.SessionChoice
	var names []string
	collect := func(choices []mcpserver.SessionChoice) {
		for _, choice := range choices {
			names = append(names, choice.Name)
			if matches(choice) {
				found = append(found, choice)
			}
		}
	}
	listed := s.providerChoices(ctx)
	for i, source := range s.sources {
		if source.IsFiles() {
			choices, err := s.fileChoices()
			if err != nil {
				return mcpserver.SessionChoice{}, err
			}
			collect(choices)
		} else {
			collect(listed[i].choices)
		}
	}
	switch len(found) {
	case 1:
		return found[0], nil
	case 0:
		return mcpserver.SessionChoice{}, fmt.Errorf("unknown session %q. Available sessions: %s", name, strings.Join(names, ", "))
	default:
		return mcpserver.SessionChoice{}, fmt.Errorf("the session name %q is ambiguous (%d sessions match)", name, len(found))
	}
}

// switchTo switches to the named session without the browser if it needs no passphrase
// or two-factor code, and otherwise opens the picker at its unlock step. The mode (if
// given) is applied to the session. Without a session, only the mode of the current
// session changes
func (s *stdioSessions) switchTo(ctx context.Context, name string, modeName string) (string, error) {
	mode := config.SessionModeUnset
	if modeName != "" {
		var err error
		if mode, err = mcpserver.ParseMode(modeName); err != nil {
			return "", err
		}
	}
	if name == "" {
		change, err := s.server.SetMode(mode, "prompt")
		if err != nil {
			return "", err
		}
		return change.Message(), nil
	}
	choice, err := s.findChoice(ctx, name)
	if err != nil {
		return "", err
	}
	if !s.mu.TryLock() {
		return "A Cumulocity session picker is already open in the browser. Wait for me to select a session.", nil
	}
	apply := func(session *pendingSession) error {
		s.use(session)
		if modeName == "" {
			return nil
		}
		_, err := s.server.SetMode(mode, "prompt")
		return err
	}
	if !choice.NeedsPassphrase && !choice.NeedsCode {
		value, err := s.login(ctx, choice, mcpserver.LoginSecrets{})
		switch {
		case err == nil:
			defer s.mu.Unlock()
			if err := apply(value.(*pendingSession)); err != nil {
				return "", err
			}
			info, runner := s.server.Active()
			return fmt.Sprintf("Switched the Cumulocity session to %s (tenant %s, user %s, mode %s). "+
				"Call %s to confirm. Don't change any data as part of this.",
				info.SessionName, info.Tenant, info.Username, runner.Mode, mcpserver.ToolSession), nil
		case errors.Is(err, mcpserver.ErrTOTPRequired):
			// ask for the code in the browser
		default:
			s.mu.Unlock()
			return "", err
		}
	}
	go func() {
		defer s.mu.Unlock()
		selection, err := s.pickerWithLogin().Unlock(context.Background(), "Unlock the Cumulocity session", "", choice)
		if err == nil {
			err = apply(selection.Value.(*pendingSession))
		}
		if err != nil && !errors.Is(err, mcpserver.ErrPickerCancelled) {
			s.log.Warnf("Session was not switched. %s", err)
		}
	}()
	return fmt.Sprintf("I'm switching the Cumulocity session to %s. It needs a passphrase or code, so I opened the browser window to unlock it. "+
		"Once I've unlocked it, call %s and tell me which tenant, user and mode are now used. Don't change any data as part of this.",
		choice.Name, mcpserver.ToolSession), nil
}
