package mcpserver

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cli/browser"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

// DefaultPickerTimeout how long the user has to select a session in the browser
const DefaultPickerTimeout = 3 * time.Minute

// ErrPickerCancelled the user cancelled the session selection
var ErrPickerCancelled = errors.New("the user cancelled the session selection")

// ErrPickerTimeout the user did not select a session in time
var ErrPickerTimeout = errors.New("the user did not select a session in time")

const pickerCookie = "c8y_session_picker"

// SessionChoice a session the user can select
type SessionChoice struct {
	// Path session file
	Path     string
	Name     string
	Host     string
	Tenant   string
	Username string

	// Current true if it is the session which is currently used
	Current bool

	// Mode the mode the agent can use with the session, e.g. "dev", and a description
	// of what it allows
	Mode     string
	ModeInfo string

	// SessionMode the session's own mode, and MaxMode the highest mode the user can choose
	// (the ceiling of the server). The mode control is only shown if MaxMode is set. Modes
	// above SessionMode are an override, and are marked as such. Mode is preselected
	SessionMode string
	MaxMode     string

	// NeedsPassphrase true if the passphrase must be entered on the page to unlock the
	// session (it is encrypted and no pin entry command is configured)
	NeedsPassphrase bool

	// NeedsCode true if a two-factor authentication code is expected, e.g. because the
	// last login needed one. The code field is also shown when the login asks for it
	NeedsCode bool

	// Source where the session comes from, shown as the heading of its group, e.g.
	// "Bitwarden". Empty for session files
	Source string

	// Provider extension session provider which listed the session, and URI its sessionUri.
	// Both are empty for session files
	Provider string
	URI      string

	// Group optional grouping within the source, e.g. the folder of a vault
	Group string
}

// ModeOption a mode in the mode control of a session
type ModeOption struct {
	Name string
	Info string
	// Disabled the mode is above the ceiling
	Disabled bool
	// Override the mode is above the session's own mode
	Override bool
	Checked  bool
}

// ModeOptions returns the modes the user can choose for the session: prod, qual and dev,
// with the ones above MaxMode disabled. It returns nothing if the session has no MaxMode
func (c SessionChoice) ModeOptions() []ModeOption {
	maxMode, err := ParseMode(c.MaxMode)
	if err != nil {
		return nil
	}
	sessionMode := SessionDefaultMode(config.SessionModeUnset.FromString(c.SessionMode, false))
	current, _ := ParseMode(c.Mode)
	options := []ModeOption{}
	for _, m := range []config.SessionMode{config.SessionModeProduction, config.SessionModeQual, config.SessionModeDev} {
		options = append(options, ModeOption{
			Name:     m.String(),
			Info:     ModeInfo(m),
			Disabled: m > maxMode,
			Override: m > sessionMode,
			Checked:  m == current,
		})
	}
	return options
}

// SourceError a source of sessions (e.g. a session provider) which could not list its
// sessions. It is shown in the picker instead of the source's sessions
type SourceError struct {
	Source  string
	Message string
}

// SessionFilesSource heading of the sessions from session files
const SessionFilesSource = "Session files"

// LoginFunc logs in to the selected session with the secrets entered on the page. The
// passphrase is only set for sessions which need it (see SessionChoice.NeedsPassphrase).
// It returns ErrTOTPRequired if a two-factor code is needed, and a value which is
// returned by Pick
type LoginFunc func(ctx context.Context, choice SessionChoice, secrets LoginSecrets) (value any, err error)

// Selection the session selected by the user
type Selection struct {
	Choice SessionChoice
	// Mode the mode the user chose (a mode up to the ceiling), or unset if the page had no
	// mode control. The caller applies it
	Mode config.SessionMode
	// ModeOnly true if only the mode of the current session is changed: no login was done
	// and Value is nil
	ModeOnly bool
	// Value returned by the login (if any)
	Value any
}

// WebPicker lets the user select a session in the browser. It is used when the
// server has no terminal (stdio mode). A temporary web server listens on the
// loopback interface for a single selection. As other local processes can reach
// loopback ports, the page can only be opened with a random one-time code (passed
// in the URL which is opened in the browser), which is exchanged for a cookie, and
// the selection must come from the page itself (origin, host and csrf checks).
type WebPicker struct {
	// OpenBrowser opens the url in the browser. Defaults to the system browser
	OpenBrowser func(url string) error

	// Notify optional callback to draw the user's attention to the browser
	Notify func(message string)

	// Timeout how long the user has to select a session. Defaults to DefaultPickerTimeout
	Timeout time.Duration

	// Login optional function which logs in to the selected session, so the user sees
	// straight away whether the session works, and can retry (e.g. after entering a
	// wrong passphrase). The selection is only finished when the login succeeds
	Login LoginFunc
}

func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func equalToken(a, b string) bool {
	return a != "" && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

type pickerResult struct {
	selection *Selection
	err       error
}

// Pick shows the sessions in the browser and returns the one the user selected (and
// logged in to, if Login is set)
func (w *WebPicker) Pick(ctx context.Context, title string, choices []SessionChoice, problems ...SourceError) (*Selection, error) {
	if len(choices) == 0 {
		if len(problems) > 0 {
			messages := make([]string, 0, len(problems))
			for _, p := range problems {
				messages = append(messages, p.Source+": "+p.Message)
			}
			return nil, fmt.Errorf("no sessions were found. %s", strings.Join(messages, "; "))
		}
		return nil, errors.New("no sessions were found. Create a session with \"c8y sessions create\" first")
	}
	return w.run(ctx, title, "", choices, problems, false)
}

// pickerGroup sessions of one source shown in the picker
type pickerGroup struct {
	Label string
	Error string
	Items []pickerItem
}

// pickerItem a session and its index in the list of choices
type pickerItem struct {
	Index  int
	Choice SessionChoice
}

// ModeOptions the modes of the session, see SessionChoice.ModeOptions
func (i pickerItem) ModeOptions() []ModeOption { return i.Choice.ModeOptions() }

// showGroupHeadings returns true if the headings of the groups are shown: if there is more than
// one source, a source has an error, or a session provider is used, so the user can see where
// the sessions come from. Only session files are shown without a heading
func showGroupHeadings(groups []pickerGroup) bool {
	if len(groups) > 1 {
		return true
	}
	for _, g := range groups {
		if g.Error != "" || g.Label != SessionFilesSource {
			return true
		}
	}
	return false
}

// groupChoices groups the sessions by their source, in the order of the choices. Sources
// which could not list their sessions are added with their error
func groupChoices(choices []SessionChoice, problems []SourceError) []pickerGroup {
	groups := []pickerGroup{}
	find := func(label string) int {
		for i := range groups {
			if groups[i].Label == label {
				return i
			}
		}
		groups = append(groups, pickerGroup{Label: label})
		return len(groups) - 1
	}
	for i, c := range choices {
		label := c.Source
		if label == "" {
			label = SessionFilesSource
		}
		g := find(label)
		groups[g].Items = append(groups[g].Items, pickerItem{Index: i, Choice: c})
	}
	for _, p := range problems {
		g := find(p.Source)
		groups[g].Error = p.Message
	}
	return groups
}

// Unlock asks the user to log in to a single session again, e.g. to enter the
// passphrase when the token expires
func (w *WebPicker) Unlock(ctx context.Context, title string, message string, choice SessionChoice) (*Selection, error) {
	return w.run(ctx, title, message, []SessionChoice{choice}, nil, true)
}

// pickerView data of the page
type pickerView struct {
	Title   string
	Message string
	Nonce   string
	CSRF    string
	Timeout int
	Choices []SessionChoice
	// Groups the choices grouped by source. ShowGroups is true if there is more than
	// one source (or a source has an error), so the headings are shown
	Groups     []pickerGroup
	ShowGroups bool
	CanGoBack  bool

	// Unlock the session for which the passphrase or code is entered
	Unlock      *SessionChoice
	UnlockIndex int
	// AskPassphrase, AskCode which fields are shown
	AskPassphrase bool
	AskCode       bool

	// Mode the mode chosen for the session to unlock, posted again with the secrets
	Mode string

	Error     string
	Done      *SessionChoice
	ModeOnly  bool
	Cancelled bool
}

func (w *WebPicker) run(ctx context.Context, title string, message string, choices []SessionChoice, problems []SourceError, unlockOnly bool) (*Selection, error) {
	timeout := w.Timeout
	if timeout <= 0 {
		timeout = DefaultPickerTimeout
	}
	openBrowser := w.OpenBrowser
	if openBrowser == nil {
		openBrowser = browser.OpenURL
	}

	// cancelled when the picker finishes, so a login which is still running doesn't
	// change anything afterwards
	runCtx, cancelRun := context.WithCancel(ctx)
	defer cancelRun()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("could not start the session picker. %w", err)
	}
	host := listener.Addr().String()
	origin := "http://" + host

	code := randomToken()
	cookieValue := randomToken()
	csrf := randomToken()
	var mu sync.Mutex
	codeUsed := false
	finished := false
	// only one login at a time
	var loginMu sync.Mutex
	// sessions which need a two-factor code, and the passphrase which was entered for
	// them, so it doesn't have to be entered again with the code. Kept until the
	// picker finishes
	needsCode := map[int]bool{}
	pendingPassphrase := map[int]string{}
	defer func() {
		mu.Lock()
		clear(pendingPassphrase)
		mu.Unlock()
	}()
	results := make(chan pickerResult, 1)
	setHeaders := func(rw http.ResponseWriter, nonce string) {
		h := rw.Header()
		h.Set("Cache-Control", "no-store")
		// not "no-referrer": browsers then send "Origin: null" with the form posts, which
		// the origin check rejects. The page's address is still never sent to other sites
		h.Set("Referrer-Policy", "same-origin")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", fmt.Sprintf("default-src 'none'; style-src 'nonce-%s'; script-src 'nonce-%s'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'", nonce, nonce))
	}
	finish := func(r pickerResult) {
		mu.Lock()
		finished = true
		mu.Unlock()
		select {
		case results <- r:
		default:
		}
	}
	render := func(rw http.ResponseWriter, status int, view pickerView) {
		view.Nonce = randomToken()
		view.Title = title
		view.CSRF = csrf
		view.Timeout = int(timeout.Minutes())
		view.CanGoBack = !unlockOnly
		if view.Message == "" {
			view.Message = message
		}
		setHeaders(rw, view.Nonce)
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(status)
		_ = pickerPage.Execute(rw, view)
	}
	unlockViewFor := func(index int, mode string, errorMessage string) pickerView {
		// the mode control is not shown while unlocking, but the chosen mode is
		unlockChoice := choices[index]
		unlockChoice.MaxMode = ""
		if m, err := ParseMode(mode); err == nil {
			unlockChoice.Mode = m.String()
			unlockChoice.ModeInfo = ModeInfo(m)
		}
		mu.Lock()
		_, hasPassphrase := pendingPassphrase[index]
		askCode := choices[index].NeedsCode || needsCode[index]
		mu.Unlock()
		return pickerView{
			Unlock:        &unlockChoice,
			UnlockIndex:   index,
			Mode:          mode,
			AskPassphrase: choices[index].NeedsPassphrase && !hasPassphrase,
			AskCode:       askCode,
			Error:         errorMessage,
		}
	}
	listView := func() pickerView {
		if unlockOnly {
			return unlockViewFor(0, "", "")
		}
		groups := groupChoices(choices, problems)
		return pickerView{Choices: choices, Groups: groups, ShowGroups: showGroupHeadings(groups)}
	}

	// checkRequest rejects requests which don't come from the page opened by the user
	checkRequest := func(rw http.ResponseWriter, r *http.Request, needCookie bool, needOrigin bool) bool {
		// DNS rebinding protection
		if r.Host != host {
			http.Error(rw, "invalid host", http.StatusForbidden)
			return false
		}
		if needCookie {
			c, err := r.Cookie(pickerCookie)
			if err != nil || !equalToken(c.Value, cookieValue) {
				http.Error(rw, "This page can only be used from the browser window which was opened by c8y.", http.StatusForbidden)
				return false
			}
		}
		if needOrigin && r.Header.Get("Origin") != origin {
			http.Error(rw, "invalid origin", http.StatusForbidden)
			return false
		}
		return true
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(rw http.ResponseWriter, r *http.Request) {
		if !checkRequest(rw, r, false, false) {
			return
		}
		mu.Lock()
		valid := !codeUsed && equalToken(r.URL.Query().Get("code"), code)
		if valid {
			codeUsed = true
		}
		mu.Unlock()
		if !valid {
			setHeaders(rw, randomToken())
			http.Error(rw, "This link has already been used or is invalid. Ask the agent to use the session again to open a new link.", http.StatusForbidden)
			return
		}
		http.SetCookie(rw, &http.Cookie{
			Name:     pickerCookie,
			Value:    cookieValue,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		// remove the code from the address bar
		http.Redirect(rw, r, "/pick", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /pick", func(rw http.ResponseWriter, r *http.Request) {
		if !checkRequest(rw, r, true, false) {
			return
		}
		render(rw, http.StatusOK, listView())
	})
	mux.HandleFunc("POST /select", func(rw http.ResponseWriter, r *http.Request) {
		if !checkRequest(rw, r, true, true) {
			return
		}
		if !equalToken(r.PostFormValue("csrf"), csrf) {
			http.Error(rw, "invalid request", http.StatusForbidden)
			return
		}
		var selected *SessionChoice
		var index int
		// "change mode only" posts the index of the current session as modeonly
		modeOnly := r.PostFormValue("modeonly") != ""
		indexValue := r.PostFormValue("session")
		if modeOnly {
			indexValue = r.PostFormValue("modeonly")
		}
		if _, err := fmt.Sscanf(indexValue, "%d", &index); err == nil && index >= 0 && index < len(choices) {
			selected = &choices[index]
		}
		if selected == nil {
			http.Error(rw, "unknown session", http.StatusBadRequest)
			return
		}

		// the mode is validated here, whatever the page offered: it must be one of the
		// modes of the session, up to the ceiling
		mode := config.SessionModeUnset
		modeName := r.PostFormValue(fmt.Sprintf("mode-%d", index))
		if modeName == "" && modeOnly {
			http.Error(rw, "no mode was selected", http.StatusBadRequest)
			return
		}
		if modeName != "" {
			allowed := false
			if m, err := ParseMode(modeName); err == nil {
				for _, o := range selected.ModeOptions() {
					if o.Name == modeName && !o.Disabled {
						allowed, mode = true, m
					}
				}
			}
			if !allowed {
				http.Error(rw, "the mode is not allowed", http.StatusBadRequest)
				return
			}
		}
		if modeOnly {
			if unlockOnly || !selected.Current {
				http.Error(rw, "the mode can only be changed for the current session", http.StatusBadRequest)
				return
			}
			mu.Lock()
			done := finished
			mu.Unlock()
			if done {
				http.Error(rw, "The session was already selected. You can close this tab.", http.StatusConflict)
				return
			}
			doneChoice := *selected
			doneChoice.Current = false
			doneChoice.MaxMode = ""
			doneChoice.Mode = mode.String()
			doneChoice.ModeInfo = ModeInfo(mode)
			render(rw, http.StatusOK, pickerView{Done: &doneChoice, ModeOnly: true})
			finish(pickerResult{selection: &Selection{Choice: *selected, Mode: mode, ModeOnly: true}})
			return
		}
		unlockView := func(errorMessage string) pickerView {
			return unlockViewFor(index, modeName, errorMessage)
		}
		errorView := func(errorMessage string) pickerView {
			mu.Lock()
			askCode := needsCode[index]
			mu.Unlock()
			if selected.NeedsPassphrase || selected.NeedsCode || askCode || unlockOnly {
				return unlockView(errorMessage)
			}
			v := listView()
			v.Error = errorMessage
			return v
		}

		// the secrets entered on the page
		secrets := LoginSecrets{TOTPCode: strings.ReplaceAll(r.PostFormValue("code"), " ", "")}
		mu.Lock()
		stored, hasStored := pendingPassphrase[index]
		askCode := selected.NeedsCode || needsCode[index]
		mu.Unlock()
		if selected.NeedsPassphrase {
			secrets.Passphrase = r.PostFormValue("passphrase")
			if secrets.Passphrase == "" && hasStored {
				secrets.Passphrase = stored
			}
		}
		// show the fields first, if input is needed
		missing := ""
		if selected.NeedsPassphrase && secrets.Passphrase == "" {
			missing = "Enter the passphrase."
		} else if askCode && secrets.TOTPCode == "" {
			missing = "Enter the code from your authenticator app."
		}
		if missing != "" {
			if r.PostFormValue("unlock") == "" {
				missing = ""
			}
			render(rw, http.StatusOK, unlockView(missing))
			return
		}

		if !loginMu.TryLock() {
			render(rw, http.StatusConflict, errorView("A login is already in progress. Wait for it to finish."))
			return
		}
		defer loginMu.Unlock()
		mu.Lock()
		done := finished
		mu.Unlock()
		if done {
			http.Error(rw, "The session was already selected. You can close this tab.", http.StatusConflict)
			return
		}

		var value any
		if w.Login != nil {
			// stop the login when the picker finishes or the browser disconnects
			loginCtx, cancelLogin := context.WithCancel(runCtx)
			stop := context.AfterFunc(r.Context(), cancelLogin)
			var err error
			value, err = w.Login(loginCtx, *selected, secrets)
			stop()
			cancelLogin()
			if errors.Is(err, ErrTOTPRequired) {
				mu.Lock()
				needsCode[index] = true
				if secrets.Passphrase != "" {
					pendingPassphrase[index] = secrets.Passphrase
				}
				mu.Unlock()
				msg := ""
				if secrets.TOTPCode != "" {
					msg = "The code was not accepted. Enter the current code from your authenticator app."
				}
				render(rw, http.StatusOK, unlockView(msg))
				return
			}
			if err != nil {
				render(rw, http.StatusOK, errorView(err.Error()))
				return
			}
			if runCtx.Err() != nil {
				http.Error(rw, "The session picker has expired.", http.StatusGone)
				return
			}
		}
		// "current" refers to the session before the selection
		doneChoice := *selected
		doneChoice.Current = false
		doneChoice.MaxMode = ""
		if mode != config.SessionModeUnset {
			doneChoice.Mode = mode.String()
			doneChoice.ModeInfo = ModeInfo(mode)
		}
		render(rw, http.StatusOK, pickerView{Done: &doneChoice})
		finish(pickerResult{selection: &Selection{Choice: *selected, Mode: mode, Value: value}})
	})
	mux.HandleFunc("POST /cancel", func(rw http.ResponseWriter, r *http.Request) {
		if !checkRequest(rw, r, true, true) {
			return
		}
		if !equalToken(r.PostFormValue("csrf"), csrf) {
			http.Error(rw, "invalid request", http.StatusForbidden)
			return
		}
		render(rw, http.StatusOK, pickerView{Cancelled: true})
		finish(pickerResult{err: ErrPickerCancelled})
	})

	server := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = server.Serve(listener) }()
	defer func() {
		// let the last page finish loading before stopping
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if err := openBrowser(origin + "/?code=" + code); err != nil {
		return nil, fmt.Errorf("could not open the session picker in the browser. %w", err)
	}
	if w.Notify != nil {
		notice := "Select the Cumulocity session for the MCP server in your browser."
		if unlockOnly {
			notice = "Log in to the Cumulocity session for the MCP server in your browser."
		}
		go w.Notify(notice)
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case r := <-results:
		return r.selection, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, ErrPickerTimeout
	}
}

// pickerPage follows the style of the page shown after an OAuth2 browser login
// (go-c8y browser_success.html). The wordmark is the official Cumulocity logo
// (CC-BY-SA-4.0, Cumulocity GmbH), https://commons.wikimedia.org/wiki/File:C8y-logo-yb-cmyk.svg
var pickerPage = template.Must(template.New("picker").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="dark">
<title>{{.Title}} - Cumulocity</title>
<style nonce="{{.Nonce}}">
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  :root {
    --bg: #0d1117; --fg: #e8f0fe; --strong: #ffffff; --muted: rgba(255,255,255,.55); --faint: rgba(255,255,255,.4);
    --surface: rgba(255,255,255,.05); --surface-2: rgba(255,255,255,.04); --hover: rgba(255,255,255,.08);
    --line: rgba(255,255,255,.1); --line-2: rgba(255,255,255,.06);
    --accent: #ffd200; --accent-soft: rgba(255,210,0,.1); --accent-line: rgba(255,210,0,.35);
    --warn: #ffb74d; --warn-soft: rgba(255,183,77,.12); --warn-line: rgba(255,183,77,.45);
    --error: #ff8a80; --error-soft: rgba(255,82,82,.1); --error-line: rgba(255,82,82,.4);
    --mono: ui-monospace, "Cascadia Code", "SF Mono", Menlo, monospace;
  }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    background: var(--bg); color: var(--fg);
    min-height: 100vh; display: flex; align-items: center; justify-content: center;
    padding: 32px 16px;
  }
  body::before {
    content: ""; position: fixed; inset: 0; pointer-events: none;
    background:
      radial-gradient(1px 1px at 10% 20%, rgba(255,255,255,.5) 0%, transparent 100%),
      radial-gradient(1px 1px at 30% 65%, rgba(255,255,255,.4) 0%, transparent 100%),
      radial-gradient(1px 1px at 55% 10%, rgba(255,255,255,.5) 0%, transparent 100%),
      radial-gradient(1px 1px at 75% 45%, rgba(255,255,255,.4) 0%, transparent 100%),
      radial-gradient(1px 1px at 88% 80%, rgba(255,255,255,.5) 0%, transparent 100%),
      radial-gradient(1px 1px at 20% 88%, rgba(255,255,255,.3) 0%, transparent 100%),
      radial-gradient(1px 1px at 65% 72%, rgba(255,255,255,.3) 0%, transparent 100%);
    animation: twinkle 6s ease-in-out infinite alternate;
  }
  @keyframes twinkle { from { opacity: .6; } to { opacity: 1; } }
  @keyframes rise { from { opacity: 0; transform: translateY(32px); } to { opacity: 1; transform: translateY(0); } }
  @keyframes fade-in { from { opacity: 0; } to { opacity: 1; } }
  @keyframes spin { to { transform: rotate(360deg); } }

  .card {
    position: relative; width: 100%; max-width: 560px;
    background: var(--surface); border: 1px solid var(--line); border-radius: 20px;
    padding: 40px 40px 32px; backdrop-filter: blur(16px);
    animation: rise .6s cubic-bezier(.22,1,.36,1) both;
  }
  .card.center { max-width: 480px; text-align: center; padding: 56px 48px 48px; }
  .logo { display: block; width: 160px; height: auto; fill: #ffffff; margin-bottom: 28px; animation: fade-in .8s .1s both; }
  .center .logo { width: 220px; margin: 0 auto 36px; }
  h1 { font-size: 1.5rem; font-weight: 700; color: var(--strong); margin-bottom: 8px; }
  p { font-size: .95rem; color: var(--muted); line-height: 1.6; }
  .lead { margin-bottom: 24px; }
  .label {
    display: block; margin: 0 0 8px; font-size: .75rem; font-weight: 600; letter-spacing: .06em;
    text-transform: uppercase; color: var(--faint);
  }

  .field + .field { margin-top: 16px; }
  input.code { font-family: var(--mono); font-size: 1.1rem; letter-spacing: .3em; }
  input[type=search], input[type=password], input.code {
    width: 100%; padding: 12px 14px; font: inherit; font-size: .95rem; color: var(--strong);
    background: var(--surface-2); border: 1px solid var(--line); border-radius: 12px; outline: none;
    transition: border-color .15s, box-shadow .15s;
  }
  input::placeholder { color: var(--faint); }
  input:focus { border-color: var(--accent-line); box-shadow: 0 0 0 3px var(--accent-soft); }
  input[type=search] { margin-bottom: 12px; }

  .sessions {
    list-style: none; background: var(--surface-2); border: 1px solid var(--line-2); border-radius: 12px;
    max-height: min(52vh, 460px); overflow-y: auto;
  }
  .sessions li + li { border-top: 1px solid var(--line-2); }
  .sessions li.group {
    /* the source is shown as written, e.g. the provider name */
    position: sticky; top: 0; z-index: 1; padding: 8px 16px; font-size: .78rem; font-weight: 600; color: var(--faint);
    /* opaque (the list's colour on the card), as the sessions scroll underneath the heading */
    background: #22262b; border-bottom: 1px solid var(--line-2);
  }
  .sessions li.problem { padding: 10px 16px; font-size: .85rem; color: var(--error); overflow-wrap: anywhere; }
  .session {
    width: 100%; display: grid; gap: 4px; text-align: left; padding: 12px 16px; cursor: pointer;
    background: none; border: 0; color: inherit; font: inherit; transition: background .15s;
  }
  .session:hover, .session:focus-visible { background: var(--hover); outline: none; }
  .session:focus-visible { box-shadow: inset 3px 0 0 var(--accent); }
  .row { display: flex; flex-wrap: wrap; align-items: center; gap: 6px 8px; }
  .name { font-weight: 600; color: var(--strong); overflow-wrap: anywhere; }
  .details { font-family: var(--mono); font-size: .78rem; color: var(--muted); overflow-wrap: anywhere; }
  .pill {
    display: inline-block; font-size: .68rem; font-weight: 600; letter-spacing: .08em; text-transform: uppercase;
    padding: 2px 10px; border-radius: 99px; border: 1px solid var(--line); color: var(--muted);
  }
  .pill.accent { background: var(--accent-soft); border-color: var(--accent-line); color: var(--accent); }
  .mode { font-size: .78rem; color: var(--faint); }
  .modes { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 6px; padding: 0 16px 12px; }
  .modes input { position: absolute; opacity: 0; pointer-events: none; }
  .seg {
    padding: 4px 14px; font-size: .78rem; font-weight: 600; letter-spacing: .04em; cursor: pointer;
    color: var(--muted); border: 1px solid var(--line); border-radius: 99px; transition: background .15s, border-color .15s, color .15s;
  }
  .seg:hover { color: var(--strong); border-color: var(--faint); }
  .modes input:checked + .seg { background: var(--accent-soft); border-color: var(--accent-line); color: var(--accent); }
  .modes input:checked + .seg.override { background: var(--warn-soft); border-color: var(--warn-line); color: var(--warn); }
  .modes input:focus-visible + .seg { box-shadow: 0 0 0 3px var(--accent-soft); }
  .modes input:disabled + .seg { opacity: .35; cursor: not-allowed; text-decoration: line-through; }
  .modes input:disabled + .seg:hover { color: var(--muted); border-color: var(--line); }
  .can { display: none; order: 5; flex-basis: 100%; font-size: .78rem; color: var(--faint); }
  .can.override { color: var(--warn); }
  .m-prod:checked ~ .can-prod, .m-qual:checked ~ .can-qual, .m-dev:checked ~ .can-dev { display: block; }
  .button.small { order: 6; margin-left: auto; padding: 4px 12px; font-size: .78rem; }
  .empty { padding: 14px 16px; color: var(--faint); font-size: .9rem; display: none; }

  kbd {
    display: inline-block; min-width: 1.6em; padding: 1px 6px; font-family: var(--mono); font-size: .72rem; line-height: 1.5;
    text-align: center; color: var(--fg); background: var(--surface); border: 1px solid var(--line); border-bottom-width: 2px; border-radius: 6px;
  }
  .keys { display: flex; flex-wrap: wrap; align-items: center; gap: 6px 14px; margin-top: 12px; font-size: .78rem; color: var(--faint); }
  .keys[hidden] { display: none; }
  .key-modes[hidden] { display: none !important; }
  .keys span { display: inline-flex; align-items: center; gap: 4px; }
  .link { padding: 0; font: inherit; color: var(--muted); background: none; border: 0; text-decoration: underline; text-underline-offset: 3px; cursor: pointer; }
  .link:hover, .link:focus-visible { color: var(--strong); outline: none; }
  dialog.help {
    margin: auto; width: min(480px, calc(100vw - 32px)); padding: 28px 28px 24px; color: var(--fg);
    background: #161b22; border: 1px solid var(--line); border-radius: 16px;
  }
  dialog.help::backdrop { background: rgba(0,0,0,.6); backdrop-filter: blur(2px); }
  dialog.help h2 { font-size: 1.1rem; font-weight: 700; color: var(--strong); margin-bottom: 16px; }
  dialog.help h3 { margin: 20px 0 8px; font-size: .75rem; font-weight: 600; letter-spacing: .06em; text-transform: uppercase; color: var(--faint); }
  dialog.help dl { display: grid; grid-template-columns: max-content 1fr; gap: 10px 16px; align-items: baseline; font-size: .88rem; }
  dialog.help dt { display: flex; flex-wrap: wrap; gap: 4px; }
  dialog.help dd { color: var(--muted); line-height: 1.5; }
  dialog.help code { font-family: var(--mono); font-size: .82rem; color: var(--fg); }
  dialog.help .actions { margin-top: 24px; }
  dialog.help .note { margin-top: 12px; font-size: .82rem; }

  .selected {
    display: grid; gap: 4px; padding: 14px 16px; margin-bottom: 20px;
    background: var(--surface-2); border: 1px solid var(--line-2); border-radius: 12px;
  }
  .error {
    margin-bottom: 16px; padding: 12px 14px; font-size: .9rem; color: var(--error);
    background: var(--error-soft); border: 1px solid var(--error-line); border-radius: 12px; overflow-wrap: anywhere;
  }
  .actions { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8px; margin-top: 20px; }
  .actions form { display: contents; }
  .button {
    display: inline-flex; align-items: center; gap: 8px; padding: 10px 18px; font: inherit; font-size: .9rem; font-weight: 600;
    border-radius: 10px; cursor: pointer; text-decoration: none; transition: background .15s, border-color .15s, color .15s;
    background: none; color: var(--muted); border: 1px solid var(--line);
  }
  .button:hover, .button:focus-visible { color: var(--strong); border-color: rgba(255,255,255,.35); outline: none; }
  .button.primary { background: var(--accent); border-color: var(--accent); color: #0d1117; }
  .button.primary:hover, .button.primary:focus-visible { background: #ffdc33; border-color: #ffdc33; color: #0d1117; }
  .status { display: none; align-items: center; justify-content: center; gap: 10px; margin-top: 20px; font-size: .9rem; color: var(--muted); text-align: center; }
  .spinner { flex-shrink: 0; width: 16px; height: 16px; border-radius: 50%; border: 2px solid var(--accent-line); border-top-color: var(--accent); animation: spin .8s linear infinite; }
  body.busy .status { display: flex; }
  body.busy .button, body.busy .session, body.busy input { opacity: .45; pointer-events: none; }

  .check-ring { width: 80px; height: 80px; margin: 0 auto 28px; display: block; }
  .check-ring circle {
    fill: none; stroke: var(--accent); stroke-width: 4; stroke-linecap: round;
    stroke-dasharray: 220; stroke-dashoffset: 220; transform-origin: center; transform: rotate(-90deg);
    animation: draw .7s .3s cubic-bezier(.65,0,.35,1) forwards;
  }
  .check-ring polyline {
    fill: none; stroke: var(--accent); stroke-width: 5; stroke-linecap: round; stroke-linejoin: round;
    stroke-dasharray: 60; stroke-dashoffset: 60; animation: draw .4s .95s ease forwards;
  }
  @keyframes draw { to { stroke-dashoffset: 0; } }
  .center h1 { animation: fade-in .5s .8s both; }
  .center p { animation: fade-in .5s 1s both; }
  .center .summary { margin-top: 24px; text-align: left; animation: fade-in .5s 1.1s both; }
  .badge { display: inline-block; margin-top: 28px; padding: 6px 18px; font-size: .8rem; animation: fade-in .5s 1.2s both; }
  @media (max-width: 480px) {
    .card, .card.center { padding: 32px 20px 24px; border-radius: 16px; }
  }
  @media (prefers-reduced-motion: reduce) {
    *, *::before { animation: none !important; transition: none !important; }
    .check-ring circle, .check-ring polyline { stroke-dashoffset: 0; }
  }
</style>
</head>
<body>
{{define "session"}}
  <span class="row"><span class="name">{{.Name}}</span>{{if .Current}}<span class="pill accent">current</span>{{end}}{{if .NeedsPassphrase}}<span class="pill">encrypted</span>{{end}}{{if .Group}}<span class="pill">{{.Group}}</span>{{end}}</span>
  <span class="details">{{.Host}}{{if .Tenant}} · {{.Tenant}}{{end}}{{if .Username}} · {{.Username}}{{end}}</span>
  {{if and .ModeInfo (not .MaxMode)}}<span class="mode">Agent can {{.ModeInfo}} ({{.Mode}})</span>{{end}}
{{end}}
{{define "modes"}}{{$i := .Index}}
  <div class="modes" role="radiogroup" aria-label="Mode of the agent for {{.Choice.Name}}">
    {{range .ModeOptions}}<input type="radio" class="m-{{.Name}}" id="m{{$i}}-{{.Name}}" name="mode-{{$i}}" value="{{.Name}}"{{if .Checked}} checked{{end}}{{if .Disabled}} disabled{{end}} aria-keyshortcuts="{{if eq .Name "prod"}}P 1{{else if eq .Name "qual"}}Q 2{{else}}D 3{{end}}"><label for="m{{$i}}-{{.Name}}" class="seg{{if .Override}} override{{end}}" title="{{if .Disabled}}Above the maximum mode of the MCP server{{else if .Override}}Override: above the session's own mode{{else}}Agent can {{.Info}}{{end}}">{{.Name}}</label>
    {{end}}
    {{range .ModeOptions}}<span class="can can-{{.Name}}{{if .Override}} override{{end}}">Agent can {{.Info}}{{if .Override}}. Override: above the session's own mode{{end}}</span>
    {{end}}
    {{if .Choice.Current}}<button class="button small" type="submit" name="modeonly" value="{{$i}}">Change mode only</button>{{end}}
  </div>
{{end}}
{{define "logo"}}<svg class="logo" role="img" aria-label="Cumulocity" viewBox="0 0 873 189" xmlns="http://www.w3.org/2000/svg">
<g clip-path="url(#a)">
<path d="M414.671 169.084V94.444H396.384V186.728H455.309V169.084H414.671Z"/>
<path d="M508.37 92.794C481.587 92.794 460.77 113.739 460.77 140.396C460.77 167.18 481.714 187.997 508.37 187.997C534.645 187.997 555.97 166.672 555.97 140.396C555.97 114.246 534.645 92.794 508.37 92.794ZM537.437 140.523C537.437 157.278 524.744 170.48 508.37 170.48C491.995 170.48 479.175 157.405 479.175 140.523C479.175 123.767 491.995 110.565 508.37 110.565C524.744 110.692 537.437 123.767 537.437 140.523Z"/>
<path d="M48.55 110.692C59.471 110.692 68.615 115.897 74.33 125.417L75.219 126.94L90.966 117.801L90.077 116.278C81.568 101.553 66.075 92.921 48.55 92.921C21.755 92.921 0.928 113.866 0.928 140.523C0.928 167.307 21.882 188.124 48.55 188.124C66.075 188.124 81.568 179.366 90.077 164.768L90.966 163.245L75.219 154.105L74.33 155.628C68.615 165.149 59.344 170.353 48.55 170.353C32.168 170.353 19.342 157.278 19.342 140.396C19.215 123.767 32.041 110.692 48.55 110.692Z"/>
<path d="M160.432 150.424C160.432 162.864 153.447 170.48 141.764 170.48C129.953 170.48 122.969 162.991 122.969 150.424V94.444H104.682V150.17C104.682 172.892 119.667 188.252 141.891 188.252C163.987 188.252 178.973 172.892 178.973 150.17V94.444H160.686V150.424H160.432Z"/>
<path d="M361.08 150.424C361.08 162.864 354.095 170.48 342.412 170.48C330.602 170.48 323.617 162.991 323.617 150.424V94.444H305.33V150.17C305.33 172.892 320.315 188.252 342.539 188.252C364.636 188.252 379.621 172.892 379.621 150.17V94.444H361.334V150.424H361.08Z"/>
<path d="M616.717 110.692C627.638 110.692 636.782 115.897 642.497 125.417L643.386 126.94L659.133 117.801L658.244 116.278C649.735 101.553 634.242 92.921 616.717 92.921C589.922 92.921 569.095 113.866 569.095 140.523C569.095 167.307 590.049 188.124 616.717 188.124C634.242 188.124 649.735 179.366 658.244 164.768L659.133 163.245L643.386 154.105L642.497 155.628C636.782 165.149 627.512 170.353 616.717 170.353C600.335 170.353 587.509 157.278 587.509 140.396C587.382 123.767 600.335 110.692 616.717 110.692Z"/>
<path d="M693.168 94.444H674.881V186.728H693.168V94.444Z"/>
<path d="M709.041 112.089H734.821V186.728H753.108V112.089H778.887V94.444H709.041V112.089Z"/>
<path d="M850.892 94.444L828.033 129.479L805.429 94.444H783.84L818.89 148.012V186.728H837.304V148.139L872.481 94.444H850.892Z"/>
<path d="M241.834 142.173L211.99 94.444H195.735V186.728H214.022V130.495L235.992 165.53H247.93L269.772 130.622V186.728H288.059V94.444H271.804L241.834 142.173Z"/>
<path d="M405.488 63.288C422.659 63.288 436.776 49.297 436.776 32C436.776 14.83 422.786 0.712 405.488 0.712C388.191 0.712 374.2 14.702 374.2 32C374.2 49.17 388.191 63.288 405.488 63.288Z"/>
<path d="M475.245 1.791H454.376V62.209H475.245V1.791Z"/>
<path d="M524.133 63.29C541.304 63.29 555.421 49.298 555.421 32C555.421 14.829 541.431 0.71 524.133 0.71C506.836 0.71 492.845 14.701 492.845 32C492.972 49.171 506.836 63.29 524.133 63.29Z"/>
<path d="M604.31 63.29C621.48 63.29 635.598 49.298 635.598 32C635.598 14.829 621.607 0.71 604.31 0.71C587.139 0.71 573.021 14.701 573.021 32C573.149 49.171 587.139 63.29 604.31 63.29Z"/>
<path d="M684.486 63.29C701.656 63.29 715.774 49.298 715.774 32C715.774 14.829 701.783 0.71 684.486 0.71C667.188 0.71 653.198 14.701 653.198 32C653.325 49.171 667.315 63.29 684.486 63.29Z"/>
<path d="M764.662 63.29C781.832 63.29 795.95 49.298 795.95 32C795.95 14.829 781.959 0.71 764.662 0.71C747.492 0.71 733.374 14.701 733.374 32C733.374 49.171 747.364 63.29 764.662 63.29Z"/>
<path d="M834.419 1.791H813.55V62.209H834.419V1.791Z"/>
<path d="M872.888 1.791H852.019V62.209H872.888V1.791Z"/>
</g>
<defs>
<clipPath id="a"><rect width="873" height="188" transform="translate(0 0.5)"/></clipPath>
</defs>
</svg>{{end}}
{{if .Done}}
<main class="card center">
  {{template "logo"}}
  <svg class="check-ring" viewBox="0 0 80 80" aria-hidden="true"><circle cx="40" cy="40" r="34"/><polyline points="24,40 34,51 56,28"/></svg>
  <h1>{{if .ModeOnly}}Mode changed{{else}}Session selected{{end}}</h1>
  <p>{{if .ModeOnly}}The MCP server keeps using this session with the new mode.{{else}}The MCP server now uses this session.{{end}}<br>You can close this tab and return to your agent.</p>
  <div class="selected summary">{{template "session" .Done}}</div>
  <span class="pill accent badge">&#x2714;&nbsp; {{if .ModeOnly}}Mode active{{else}}Session active{{end}}</span>
</main>
{{else if .Cancelled}}
<main class="card center">
  {{template "logo"}}
  <h1>Session not changed</h1>
  <p>You can close this tab.</p>
</main>
{{else if .Unlock}}
<main class="card">
  {{template "logo"}}
  <h1>{{.Title}}</h1>
  <p class="lead">{{if .Message}}{{.Message}}{{else if and .AskPassphrase .AskCode}}The session is encrypted and needs two-factor authentication. Enter the passphrase and the code from your authenticator app.{{else if .AskPassphrase}}The session is encrypted. Enter the passphrase to unlock it.{{else if .AskCode}}The session needs two-factor authentication. Enter the code from your authenticator app.{{else}}Log in to the session.{{end}}</p>
  <div class="selected">{{template "session" .Unlock}}</div>
  {{if .Error}}<p class="error" role="alert">{{.Error}}</p>{{end}}
  <form method="post" action="/select" class="login" id="login">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <input type="hidden" name="session" value="{{.UnlockIndex}}">
    <input type="hidden" name="unlock" value="1">
    {{if .Mode}}<input type="hidden" name="mode-{{.UnlockIndex}}" value="{{.Mode}}">{{end}}
    {{if .AskPassphrase}}
    <div class="field">
      <label class="label" for="passphrase">Passphrase</label>
      <input type="password" id="passphrase" name="passphrase" autofocus autocomplete="off" required>
    </div>
    {{end}}
    {{if .AskCode}}
    <div class="field">
      <label class="label" for="code">Two-factor code</label>
      <input type="text" id="code" name="code" class="code" inputmode="numeric" autocomplete="one-time-code" pattern="[0-9 ]{6,10}" maxlength="10" placeholder="123456" required {{if not .AskPassphrase}}autofocus{{end}}>
    </div>
    {{end}}
  </form>
  <div class="actions">
    <form method="post" action="/cancel"><input type="hidden" name="csrf" value="{{.CSRF}}"><button class="button" type="submit">Cancel</button></form>
    {{if .CanGoBack}}<a class="button" href="/pick">Back</a>{{end}}
    <button class="button primary" type="submit" form="login">Log in</button>
  </div>
  <p class="status" role="status"><span class="spinner"></span><span>Logging in. If you are asked to confirm, for example with Touch ID, confirm it there.</span></p>
</main>
{{else}}
<main class="card">
  {{template "logo"}}
  <h1>{{.Title}}</h1>
  <p class="lead">Choose the session the AI agent works with. This page expires in {{.Timeout}} minutes.</p>
  {{if .Error}}<p class="error" role="alert">{{.Error}}</p>{{end}}
  <input type="search" id="filter" placeholder="Filter by name, host, tenant or user" autofocus autocomplete="off" aria-label="Filter sessions">
  <form method="post" action="/select" class="login">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <!-- disabled default button: Enter in a mode control must not select the first session -->
    <button type="submit" disabled hidden aria-hidden="true" tabindex="-1"></button>
    <ul class="sessions" id="sessions">
    {{range $g, $group := .Groups}}
      {{if $.ShowGroups}}<li class="group" data-group="{{$g}}">{{$group.Label}}</li>{{end}}
      {{if $group.Error}}<li class="problem" data-group="{{$g}}">{{$group.Error}}</li>{{end}}
      {{range $group.Items}}
      <li class="item" data-group="{{$g}}" data-search="{{.Choice.Name}} {{.Choice.Host}} {{.Choice.Tenant}} {{.Choice.Username}} {{.Choice.Group}} {{$group.Label}}">
        <button class="session" type="submit" name="session" value="{{.Index}}" aria-keyshortcuts="Enter">{{template "session" .Choice}}</button>
        {{if .ModeOptions}}{{template "modes" .}}{{end}}
      </li>
      {{end}}
    {{end}}
    </ul>
    <div class="empty" id="empty">No sessions match the filter.</div>
  </form>
  <!-- the shortcuts need the script, so the hint is only shown by it -->
  <p class="keys" id="keys" hidden>
    <span><kbd>↑</kbd><kbd>↓</kbd> move</span>
    <span class="key-modes"><kbd>P</kbd><kbd>Q</kbd><kbd>D</kbd> set mode</span>
    <span><kbd>Enter</kbd> select</span>
    <button class="link" type="button" id="show-help" aria-keyshortcuts="?"><kbd>?</kbd> all shortcuts</button>
  </p>
  <div class="actions">
    <form method="post" action="/cancel"><input type="hidden" name="csrf" value="{{.CSRF}}"><button class="button" type="submit">Cancel</button></form>
  </div>
  <p class="status" role="status"><span class="spinner"></span><span>Logging in. If you are asked to confirm, for example with Touch ID, confirm it there.</span></p>
</main>
<dialog class="help" id="help" aria-labelledby="help-title">
  <h2 id="help-title">Keyboard shortcuts</h2>
  <dl>
    <dt><kbd>↑</kbd><kbd>↓</kbd></dt><dd>Move between the sessions. <kbd>↑</kbd> on the first session goes back to the filter.</dd>
    <dt class="key-modes"><kbd>P</kbd> or <kbd>1</kbd></dt><dd class="key-modes">prod: the agent can only read data</dd>
    <dt class="key-modes"><kbd>Q</kbd> or <kbd>2</kbd></dt><dd class="key-modes">qual: the agent can also create and update</dd>
    <dt class="key-modes"><kbd>D</kbd> or <kbd>3</kbd></dt><dd class="key-modes">dev: the agent can also delete</dd>
    <dt><kbd>Enter</kbd></dt><dd>Select the focused session with the mode shown for it. In the filter, select the first match.</dd>
    <dt>Typing</dt><dd>Filter the sessions by name, host, tenant or user.</dd>
    <dt><kbd>?</kbd></dt><dd>Show this help. <kbd>Esc</kbd> closes it.</dd>
  </dl>
  <p class="key-modes note">The mode keys change the session under the focus, so press <kbd>↓</kbd> first when you are in the filter.</p>
  <h3>From your agent</h3>
  <dl>
    <dt><code>mode dev</code></dt><dd>Change the mode of the current session without this page. <code>mode reset</code> goes back to the session's own mode.</dd>
    <dt><code>switch_session name qual</code></dt><dd>Switch to a session, and optionally set its mode.</dd>
  </dl>
  <p class="note">These are prompts of the MCP server. In Claude Code, run them as <code>/mcp__&lt;server&gt;__mode</code>.</p>
  <div class="actions"><form method="dialog"><button class="button primary" type="submit" autofocus>Close</button></form></div>
</dialog>
{{end}}
{{if not (or .Done .Cancelled)}}
<script nonce="{{.Nonce}}">
  const filter = document.getElementById("filter");
  if (filter) {
    const items = Array.from(document.querySelectorAll("#sessions li.item"));
    const headings = Array.from(document.querySelectorAll("#sessions li.group, #sessions li.problem"));
    filter.addEventListener("input", () => {
      const terms = filter.value.toLowerCase().split(/\s+/).filter(Boolean);
      let shown = 0;
      const shownGroups = new Set();
      for (const item of items) {
        const match = terms.every(t => item.dataset.search.toLowerCase().includes(t));
        item.hidden = !match;
        if (match) { shown++; shownGroups.add(item.dataset.group); }
      }
      // hide the headings (and errors) of sources without matching sessions while filtering
      for (const heading of headings) {
        heading.hidden = terms.length > 0 && !shownGroups.has(heading.dataset.group);
      }
      document.getElementById("empty").style.display = shown ? "none" : "block";
    });
  }
  // keyboard: up/down move between the sessions, p/q/d or 1/2/3 set the mode of the
  // focused session, Enter selects it, "?" shows the help, and typing filters by name.
  // The page works without it
  const list = document.getElementById("sessions");
  if (list) {
    const buttons = () => Array.from(list.querySelectorAll("li.item:not([hidden]) button.session"));
    const modeKeys = { p: "prod", 1: "prod", q: "qual", 2: "qual", d: "dev", 3: "dev" };
    // the shortcuts only work with the script, so their hint is shown here
    const help = document.getElementById("help");
    const keys = document.getElementById("keys");
    if (!list.querySelector(".modes")) {
      for (const el of document.querySelectorAll(".key-modes")) el.hidden = true;
    }
    if (keys) keys.hidden = false;
    const showHelp = () => { if (help && !help.open && help.showModal) help.showModal(); };
    const showHelpButton = document.getElementById("show-help");
    if (showHelpButton) showHelpButton.addEventListener("click", showHelp);
    document.addEventListener("keydown", (e) => {
      if (e.ctrlKey || e.metaKey || e.altKey || e.isComposing) return;
      if (help && help.open) return;
      const target = e.target;
      // "?" shows the help, unless it is typed into a filter which already has text
      if (e.key === "?" && !(target === filter && filter.value)) {
        e.preventDefault();
        showHelp();
        return;
      }
      const item = target.closest ? target.closest("li.item") : null;
      const onRadio = !!(target.matches && target.matches(".modes input"));
      if (e.key === "ArrowDown" || e.key === "ArrowUp") {
        const all = buttons();
        if (!all.length) return;
        const at = item ? all.indexOf(item.querySelector("button.session")) : -1;
        e.preventDefault();
        if (e.key === "ArrowDown") {
          all[Math.min(at + 1, all.length - 1)].focus();
        } else if (at > 0) {
          all[at - 1].focus();
        } else if (filter) {
          filter.focus();
        }
        return;
      }
      if (e.key === "Enter" && (onRadio || target === filter)) {
        const first = onRadio ? item.querySelector("button.session") : buttons()[0];
        if (first) { e.preventDefault(); first.click(); }
        return;
      }
      if (e.key.length !== 1 || target === filter) return;
      const mode = modeKeys[e.key.toLowerCase()];
      if (item && mode) {
        const radio = item.querySelector('.modes input[value="' + mode + '"]');
        if (radio && !radio.disabled) { radio.checked = true; e.preventDefault(); }
        return;
      }
      if (filter && !(target instanceof HTMLInputElement)) filter.focus();
    });
  }
  // show that the login is running (it can take a moment, e.g. to confirm with the
  // pin entry), and prevent a second submission
  let busy = false;
  for (const form of document.querySelectorAll("form.login")) {
    form.addEventListener("submit", (e) => {
      if (busy) { e.preventDefault(); return; }
      busy = true;
      document.body.classList.add("busy");
    });
  }
</script>
{{end}}
</body>
</html>
`))
