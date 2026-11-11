package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"
)

var testChoices = []SessionChoice{
	{Path: "/home/me/.cumulocity/dev.json", Name: "dev", Host: "https://dev.example.com", Tenant: "t1", Username: "user1", Current: true},
	{Path: "/home/me/.cumulocity/prod.json", Name: "prod <script>", Host: "https://prod.example.com", Tenant: "t2", Username: "user2"},
}

// browserSession simulates the user's browser
type browserSession struct {
	t      *testing.T
	client *http.Client
	origin string
	csrf   string
	// referrerPolicy of the last page, which changes the Origin header sent by browsers
	referrerPolicy string
}

func newBrowser(t *testing.T) *browserSession {
	jar, _ := cookiejar.New(nil)
	return &browserSession{t: t, client: &http.Client{Jar: jar, Timeout: 5 * time.Second}}
}

func (b *browserSession) get(rawURL string) (int, string) {
	b.t.Helper()
	res, err := b.client.Get(rawURL)
	if err != nil {
		b.t.Fatal(err)
	}
	defer res.Body.Close()
	b.referrerPolicy = res.Header.Get("Referrer-Policy")
	body, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(body)
}

// open follows the link opened by c8y and reads the csrf token from the page
func (b *browserSession) open(link string) (int, string) {
	b.t.Helper()
	u, _ := url.Parse(link)
	b.origin = u.Scheme + "://" + u.Host
	status, body := b.get(link)
	if m := regexp.MustCompile(`name="csrf" value="([0-9a-f]+)"`).FindStringSubmatch(body); m != nil {
		b.csrf = m[1]
	}
	return status, body
}

func (b *browserSession) post(path string, form url.Values, header map[string]string) (int, string) {
	b.t.Helper()
	req, _ := http.NewRequest(http.MethodPost, b.origin+path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// like browsers, which send "Origin: null" with form posts from a "no-referrer" page
	if b.referrerPolicy == "no-referrer" {
		req.Header.Set("Origin", "null")
	} else {
		req.Header.Set("Origin", b.origin)
	}
	for k, v := range header {
		if v == "" {
			req.Header.Del(k)
		} else {
			req.Header.Set(k, v)
		}
	}
	res, err := b.client.Do(req)
	if err != nil {
		b.t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(body)
}

// startPicker runs Pick in the background and returns the link it opened
func startPicker(t *testing.T, picker *WebPicker) (string, chan pickerResult) {
	t.Helper()
	return startPickerWith(t, picker, testChoices)
}

func startPickerWith(t *testing.T, picker *WebPicker, choices []SessionChoice) (string, chan pickerResult) {
	t.Helper()
	links := make(chan string, 1)
	picker.OpenBrowser = func(u string) error { links <- u; return nil }
	results := make(chan pickerResult, 1)
	go func() {
		selection, err := picker.Pick(context.Background(), "Select a Cumulocity session", choices)
		results <- pickerResult{selection: selection, err: err}
	}()
	select {
	case link := <-links:
		return link, results
	case <-time.After(5 * time.Second):
		t.Fatal("browser was not opened")
		return "", nil
	}
}

func waitPicker(t *testing.T, results chan pickerResult) pickerResult {
	t.Helper()
	select {
	case r := <-results:
		return r
	case <-time.After(10 * time.Second):
		t.Fatal("picker did not finish")
		return pickerResult{}
	}
}

func TestWebPickerSelect(t *testing.T) {
	notified := make(chan string, 1)
	link, results := startPicker(t, &WebPicker{Notify: func(m string) { notified <- m }})
	if !strings.HasPrefix(link, "http://127.0.0.1:") {
		t.Fatalf("picker should only listen on the loopback interface: %s", link)
	}

	b := newBrowser(t)
	status, body := b.open(link)
	if status != http.StatusOK || b.csrf == "" {
		t.Fatalf("expected the session list, got %d: %s", status, body)
	}
	if !strings.Contains(body, "dev.example.com") || !strings.Contains(body, "prod &lt;script&gt;") || strings.Contains(body, "prod <script>") {
		t.Errorf("sessions should be listed and escaped: %s", body)
	}

	status, body = b.post("/select", url.Values{"csrf": {b.csrf}, "session": {"1"}}, nil)
	if status != http.StatusOK || !strings.Contains(body, "Session selected") {
		t.Errorf("unexpected response %d: %s", status, body)
	}
	r := waitPicker(t, results)
	if r.err != nil || r.selection == nil || r.selection.Choice.Path != "/home/me/.cumulocity/prod.json" {
		t.Errorf("unexpected result: %+v", r)
	}
	select {
	case <-notified:
	case <-time.After(time.Second):
		t.Errorf("expected a notification")
	}
}

func TestWebPickerCancel(t *testing.T) {
	link, results := startPicker(t, &WebPicker{})
	b := newBrowser(t)
	b.open(link)
	if status, _ := b.post("/cancel", url.Values{"csrf": {b.csrf}}, nil); status != http.StatusOK {
		t.Errorf("unexpected status %d", status)
	}
	if r := waitPicker(t, results); r.err != ErrPickerCancelled {
		t.Errorf("expected the selection to be cancelled, got %+v", r)
	}
}

func TestWebPickerTimeout(t *testing.T) {
	_, results := startPicker(t, &WebPicker{Timeout: 300 * time.Millisecond})
	if r := waitPicker(t, results); r.err != ErrPickerTimeout {
		t.Errorf("expected a timeout, got %+v", r)
	}
}

func TestWebPickerRejectsOtherClients(t *testing.T) {
	link, results := startPicker(t, &WebPicker{Timeout: 5 * time.Second})
	u, _ := url.Parse(link)
	origin := u.Scheme + "://" + u.Host

	// wrong code
	attacker := newBrowser(t)
	if status, _ := attacker.get(origin + "/?code=" + strings.Repeat("0", 64)); status != http.StatusForbidden {
		t.Errorf("wrong code: expected 403, got %d", status)
	}
	// DNS rebinding: a website resolving its own name to 127.0.0.1
	req, _ := http.NewRequest(http.MethodGet, link, nil)
	req.Host = "attacker.example.com"
	if res, err := http.DefaultClient.Do(req); err == nil {
		res.Body.Close()
		if res.StatusCode != http.StatusForbidden {
			t.Errorf("wrong host: expected 403, got %d", res.StatusCode)
		}
	}

	// the user opens the page; the code can't be used again
	user := newBrowser(t)
	if status, _ := user.open(link); status != http.StatusOK {
		t.Fatalf("user should be able to open the page")
	}
	if status, _ := attacker.get(link); status != http.StatusForbidden {
		t.Errorf("reused code: expected 403, got %d", status)
	}
	if status, _ := attacker.get(origin + "/pick"); status != http.StatusForbidden {
		t.Errorf("no cookie: expected 403, got %d", status)
	}

	// requests from other origins (e.g. a form on a website) or without the csrf token
	attacker.origin = origin
	cases := []struct {
		name   string
		b      *browserSession
		form   url.Values
		header map[string]string
	}{
		{"no cookie", attacker, url.Values{"csrf": {user.csrf}, "session": {"0"}}, nil},
		{"other origin", user, url.Values{"csrf": {user.csrf}, "session": {"0"}}, map[string]string{"Origin": "https://evil.example.com"}},
		{"no origin", user, url.Values{"csrf": {user.csrf}, "session": {"0"}}, map[string]string{"Origin": ""}},
		{"wrong csrf", user, url.Values{"csrf": {strings.Repeat("0", 64)}, "session": {"0"}}, nil},
		{"unknown session", user, url.Values{"csrf": {user.csrf}, "session": {"5"}}, nil},
	}
	for _, tc := range cases {
		if status, _ := tc.b.post("/select", tc.form, tc.header); status == http.StatusOK {
			t.Errorf("%s: request should be rejected", tc.name)
		}
	}
	select {
	case r := <-results:
		t.Fatalf("no session should have been selected: %+v", r)
	default:
	}

	// the user's own selection still works
	if status, _ := user.post("/select", url.Values{"csrf": {user.csrf}, "session": {"0"}}, nil); status != http.StatusOK {
		t.Errorf("user's selection should work")
	}
	if r := waitPicker(t, results); r.err != nil || r.selection.Choice.Name != "dev" {
		t.Errorf("unexpected result: %+v", r)
	}
}

func TestWebPickerLogin(t *testing.T) {
	choices := []SessionChoice{
		{Path: "/s/plain.json", Name: "plain", Host: "https://a.example.com"},
		{Path: "/s/enc.json", Name: "enc", Host: "https://b.example.com", NeedsPassphrase: true},
	}
	type call struct{ name, passphrase string }
	calls := make(chan call, 10)
	picker := &WebPicker{Login: func(ctx context.Context, choice SessionChoice, secrets LoginSecrets) (any, error) {
		calls <- call{choice.Name, secrets.Passphrase}
		if choice.NeedsPassphrase && secrets.Passphrase != "secret" {
			return nil, errors.New("wrong passphrase <b>")
		}
		return "value-" + choice.Name, nil
	}}

	t.Run("encrypted session asks for the passphrase and can retry", func(t *testing.T) {
		link, results := startPickerWith(t, picker, choices)
		b := newBrowser(t)
		_, body := b.open(link)
		if !strings.Contains(body, "encrypted") {
			t.Errorf("encrypted session should be marked")
		}
		// selecting the session shows the passphrase field, without logging in
		status, body := b.post("/select", url.Values{"csrf": {b.csrf}, "session": {"1"}}, nil)
		if status != http.StatusOK || !strings.Contains(body, `type="password"`) || len(calls) != 0 {
			t.Fatalf("expected the passphrase field, got %d: %s", status, body)
		}
		// empty passphrase
		_, body = b.post("/select", url.Values{"csrf": {b.csrf}, "session": {"1"}, "unlock": {"1"}}, nil)
		if !strings.Contains(body, "Enter the passphrase") || len(calls) != 0 {
			t.Errorf("expected a message to enter the passphrase: %s", body)
		}
		// wrong passphrase: the error is shown (escaped), and the picker stays open
		_, body = b.post("/select", url.Values{"csrf": {b.csrf}, "session": {"1"}, "unlock": {"1"}, "passphrase": {"wrong"}}, nil)
		if !strings.Contains(body, "wrong passphrase &lt;b&gt;") || !strings.Contains(body, `type="password"`) {
			t.Errorf("expected the error and the passphrase field: %s", body)
		}
		if c := <-calls; c != (call{"enc", "wrong"}) {
			t.Errorf("unexpected login: %+v", c)
		}
		select {
		case r := <-results:
			t.Fatalf("picker should still be open: %+v", r)
		default:
		}
		// correct passphrase
		_, body = b.post("/select", url.Values{"csrf": {b.csrf}, "session": {"1"}, "unlock": {"1"}, "passphrase": {"secret"}}, nil)
		if !strings.Contains(body, "Session selected") || !strings.Contains(body, "b.example.com") || strings.Contains(body, "secret") {
			t.Errorf("expected the login result: %s", body)
		}
		<-calls
		r := waitPicker(t, results)
		if r.err != nil || r.selection.Choice.Name != "enc" || r.selection.Value != "value-enc" {
			t.Errorf("unexpected result: %+v", r)
		}
	})

	t.Run("plain session logs in straight away", func(t *testing.T) {
		link, results := startPickerWith(t, picker, choices)
		b := newBrowser(t)
		b.open(link)
		// a passphrase sent for a session which doesn't need it is ignored
		_, body := b.post("/select", url.Values{"csrf": {b.csrf}, "session": {"0"}, "passphrase": {"x"}}, nil)
		if !strings.Contains(body, "Session selected") {
			t.Errorf("expected the login result: %s", body)
		}
		if c := <-calls; c != (call{"plain", ""}) {
			t.Errorf("unexpected login: %+v", c)
		}
		if r := waitPicker(t, results); r.err != nil || r.selection.Value != "value-plain" {
			t.Errorf("unexpected result: %+v", r)
		}
	})

	t.Run("unlock a single session", func(t *testing.T) {
		links := make(chan string, 1)
		p := *picker
		p.OpenBrowser = func(u string) error { links <- u; return nil }
		results := make(chan pickerResult, 1)
		go func() {
			sel, err := p.Unlock(context.Background(), "Log in again", "The token expires soon.", choices[1])
			results <- pickerResult{selection: sel, err: err}
		}()
		b := newBrowser(t)
		_, body := b.open(<-links)
		if !strings.Contains(body, "The token expires soon.") || !strings.Contains(body, `type="password"`) || strings.Contains(body, `href="/pick"`) {
			t.Fatalf("expected only the passphrase field: %s", body)
		}
		b.post("/select", url.Values{"csrf": {b.csrf}, "session": {"0"}, "unlock": {"1"}, "passphrase": {"secret"}}, nil)
		<-calls
		if r := waitPicker(t, results); r.err != nil || r.selection.Choice.Name != "enc" {
			t.Errorf("unexpected result: %+v", r)
		}
	})
}

func TestWebPickerTOTP(t *testing.T) {
	choices := []SessionChoice{
		{Path: "/s/enc.json", Name: "enc", Host: "https://b.example.com", NeedsPassphrase: true},
	}
	var logins []LoginSecrets
	picker := &WebPicker{Login: func(ctx context.Context, choice SessionChoice, secrets LoginSecrets) (any, error) {
		logins = append(logins, secrets)
		if secrets.Passphrase != "secret" {
			return nil, errors.New("the passphrase is not correct")
		}
		if secrets.TOTPCode != "123456" {
			return nil, ErrTOTPRequired
		}
		return "ok", nil
	}}
	link, results := startPickerWith(t, picker, choices)
	b := newBrowser(t)
	b.open(link)
	form := func(extra url.Values) url.Values {
		v := url.Values{"csrf": {b.csrf}, "session": {"0"}, "unlock": {"1"}}
		for k, val := range extra {
			v[k] = val
		}
		return v
	}

	// the login asks for a code: the code field is shown, and the passphrase isn't asked again
	_, body := b.post("/select", form(url.Values{"passphrase": {"secret"}}), nil)
	if !strings.Contains(body, `name="code"`) || strings.Contains(body, `type="password"`) {
		t.Fatalf("expected only the code field: %s", body)
	}
	// wrong code
	_, body = b.post("/select", form(url.Values{"code": {"000000"}}), nil)
	if !strings.Contains(body, "The code was not accepted") || !strings.Contains(body, `name="code"`) {
		t.Fatalf("expected the code to be rejected: %s", body)
	}
	// empty code
	_, body = b.post("/select", form(nil), nil)
	if !strings.Contains(body, "Enter the code") {
		t.Errorf("expected a message to enter the code: %s", body)
	}
	// the stored passphrase is used with the code (spaces are removed)
	_, body = b.post("/select", form(url.Values{"code": {"123 456"}}), nil)
	if !strings.Contains(body, "Session selected") {
		t.Fatalf("expected the login to succeed: %s", body)
	}
	if r := waitPicker(t, results); r.err != nil || r.selection.Value != "ok" {
		t.Errorf("unexpected result: %+v", r)
	}
	want := []LoginSecrets{{"secret", ""}, {"secret", "000000"}, {"secret", "123456"}}
	if fmt.Sprint(logins) != fmt.Sprint(want) {
		t.Errorf("unexpected logins: %+v", logins)
	}
}

func TestWebPickerKnownTOTP(t *testing.T) {
	// e.g. when renewing the token of a session which needed a code before
	choices := []SessionChoice{{Path: "/s/a.json", Name: "a", Host: "https://a.example.com", NeedsCode: true}}
	picker := &WebPicker{Login: func(ctx context.Context, choice SessionChoice, secrets LoginSecrets) (any, error) {
		return nil, nil
	}}
	link, results := startPickerWith(t, picker, choices)
	b := newBrowser(t)
	b.open(link)
	_, body := b.post("/select", url.Values{"csrf": {b.csrf}, "session": {"0"}}, nil)
	if !strings.Contains(body, `name="code"`) || strings.Contains(body, "Session selected") {
		t.Fatalf("expected the code field before logging in: %s", body)
	}
	b.post("/select", url.Values{"csrf": {b.csrf}, "session": {"0"}, "unlock": {"1"}, "code": {"123456"}}, nil)
	if r := waitPicker(t, results); r.err != nil {
		t.Errorf("unexpected result: %+v", r)
	}
}
