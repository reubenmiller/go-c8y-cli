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
	"sync"
	"time"

	"github.com/cli/browser"
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

	// NeedsPassphrase true if the passphrase must be entered on the page to unlock the
	// session (it is encrypted and no pin entry command is configured)
	NeedsPassphrase bool
}

// LoginFunc logs in to the selected session. The passphrase is only set for sessions
// which need it (see SessionChoice.NeedsPassphrase). It returns a value which is
// returned by Pick
type LoginFunc func(ctx context.Context, choice SessionChoice, passphrase string) (value any, err error)

// Selection the session selected by the user
type Selection struct {
	Choice SessionChoice
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
func (w *WebPicker) Pick(ctx context.Context, title string, choices []SessionChoice) (*Selection, error) {
	if len(choices) == 0 {
		return nil, errors.New("no sessions were found. Create a session with \"c8y sessions create\" first")
	}
	return w.run(ctx, title, "", choices, false)
}

// Unlock asks the user to log in to a single session again, e.g. to enter the
// passphrase when the token expires
func (w *WebPicker) Unlock(ctx context.Context, title string, message string, choice SessionChoice) (*Selection, error) {
	return w.run(ctx, title, message, []SessionChoice{choice}, true)
}

// pickerView data of the page
type pickerView struct {
	Title     string
	Message   string
	Nonce     string
	CSRF      string
	Timeout   int
	Choices   []SessionChoice
	CanGoBack bool

	// Unlock the session for which the passphrase is entered
	Unlock      *SessionChoice
	UnlockIndex int

	Error     string
	Done      *SessionChoice
	Cancelled bool
}

func (w *WebPicker) run(ctx context.Context, title string, message string, choices []SessionChoice, unlockOnly bool) (*Selection, error) {
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
	listView := func() pickerView {
		if unlockOnly {
			return pickerView{Unlock: &choices[0], UnlockIndex: 0}
		}
		return pickerView{Choices: choices}
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
		if _, err := fmt.Sscanf(r.PostFormValue("session"), "%d", &index); err == nil && index >= 0 && index < len(choices) {
			selected = &choices[index]
		}
		if selected == nil {
			http.Error(rw, "unknown session", http.StatusBadRequest)
			return
		}
		unlockView := func(errorMessage string) pickerView {
			return pickerView{Unlock: selected, UnlockIndex: index, Error: errorMessage}
		}
		errorView := func(errorMessage string) pickerView {
			if selected.NeedsPassphrase || unlockOnly {
				return unlockView(errorMessage)
			}
			v := listView()
			v.Error = errorMessage
			return v
		}

		passphrase := ""
		if selected.NeedsPassphrase {
			passphrase = r.PostFormValue("passphrase")
			if passphrase == "" {
				// show the passphrase field first
				msg := ""
				if r.PostFormValue("unlock") != "" {
					msg = "Enter the passphrase."
				}
				render(rw, http.StatusOK, unlockView(msg))
				return
			}
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
			value, err = w.Login(loginCtx, *selected, passphrase)
			stop()
			cancelLogin()
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
		render(rw, http.StatusOK, pickerView{Done: &doneChoice})
		finish(pickerResult{selection: &Selection{Choice: *selected, Value: value}})
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

  input[type=search], input[type=password] {
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
  .empty { padding: 14px 16px; color: var(--faint); font-size: .9rem; display: none; }

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
  <span class="row"><span class="name">{{.Name}}</span>{{if .Current}}<span class="pill accent">current</span>{{end}}{{if .NeedsPassphrase}}<span class="pill">encrypted</span>{{end}}</span>
  <span class="details">{{.Host}}{{if .Tenant}} · {{.Tenant}}{{end}}{{if .Username}} · {{.Username}}{{end}}</span>
  {{if .ModeInfo}}<span class="mode">Agent can {{.ModeInfo}} ({{.Mode}})</span>{{end}}
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
  <h1>Session selected</h1>
  <p>The MCP server now uses this session.<br>You can close this tab and return to your agent.</p>
  <div class="selected summary">{{template "session" .Done}}</div>
  <span class="pill accent badge">&#x2714;&nbsp; Session active</span>
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
  <p class="lead">{{if .Message}}{{.Message}}{{else if .Unlock.NeedsPassphrase}}The session is encrypted. Enter the passphrase to unlock it.{{else}}Log in to the session.{{end}}</p>
  <div class="selected">{{template "session" .Unlock}}</div>
  {{if .Error}}<p class="error" role="alert">{{.Error}}</p>{{end}}
  <form method="post" action="/select" class="login" id="login">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <input type="hidden" name="session" value="{{.UnlockIndex}}">
    <input type="hidden" name="unlock" value="1">
    {{if .Unlock.NeedsPassphrase}}
    <label class="label" for="passphrase">Passphrase</label>
    <input type="password" id="passphrase" name="passphrase" autofocus autocomplete="off" required>
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
    <ul class="sessions" id="sessions">
    {{range $i, $s := .Choices}}
      <li data-search="{{$s.Name}} {{$s.Host}} {{$s.Tenant}} {{$s.Username}}">
        <button class="session" type="submit" name="session" value="{{$i}}">{{template "session" $s}}</button>
      </li>
    {{end}}
    </ul>
    <div class="empty" id="empty">No sessions match the filter.</div>
  </form>
  <div class="actions">
    <form method="post" action="/cancel"><input type="hidden" name="csrf" value="{{.CSRF}}"><button class="button" type="submit">Cancel</button></form>
  </div>
  <p class="status" role="status"><span class="spinner"></span><span>Logging in. If you are asked to confirm, for example with Touch ID, confirm it there.</span></p>
</main>
{{end}}
{{if not (or .Done .Cancelled)}}
<script nonce="{{.Nonce}}">
  const filter = document.getElementById("filter");
  if (filter) {
    const items = Array.from(document.querySelectorAll("#sessions li"));
    filter.addEventListener("input", () => {
      const terms = filter.value.toLowerCase().split(/\s+/).filter(Boolean);
      let shown = 0;
      for (const item of items) {
        const match = terms.every(t => item.dataset.search.toLowerCase().includes(t));
        item.hidden = !match;
        if (match) shown++;
      }
      document.getElementById("empty").style.display = shown ? "none" : "block";
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
