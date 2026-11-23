package mcpserver

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestGroupChoices(t *testing.T) {
	choices := []SessionChoice{
		{Name: "dev", Path: "/s/dev.json"},
		{Name: "vault-1", Source: "bitwarden", Provider: "bitwarden", URI: "bitwarden://1"},
		{Name: "prod", Path: "/s/prod.json"},
	}
	groups := groupChoices(choices, []SourceError{{Source: "onepassword", Message: "not signed in"}})
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %+v", groups)
	}
	if groups[0].Label != SessionFilesSource || len(groups[0].Items) != 2 || groups[0].Items[1].Index != 2 {
		t.Errorf("session files should be grouped and keep their index: %+v", groups[0])
	}
	if groups[1].Label != "bitwarden" || len(groups[1].Items) != 1 || groups[1].Items[0].Index != 1 {
		t.Errorf("unexpected provider group: %+v", groups[1])
	}
	if groups[2].Label != "onepassword" || groups[2].Error != "not signed in" || len(groups[2].Items) != 0 {
		t.Errorf("unexpected error group: %+v", groups[2])
	}
}

func TestWebPickerProviderSessions(t *testing.T) {
	choices := []SessionChoice{
		testChoices[0],
		{Name: "Vault prod", Host: "https://vault.example.com", Source: "bitwarden", Provider: "bitwarden", URI: "bitwarden://item-1", Group: "c8y-folder"},
	}
	picker := &WebPicker{}
	links := make(chan string, 1)
	picker.OpenBrowser = func(u string) error { links <- u; return nil }
	results := make(chan pickerResult, 1)
	go func() {
		selection, err := picker.Pick(context.Background(), "Select", choices, SourceError{Source: "onepassword", Message: "op is not signed in"})
		results <- pickerResult{selection: selection, err: err}
	}()
	var link string
	select {
	case link = <-links:
	case <-time.After(5 * time.Second):
		t.Fatal("browser was not opened")
	}

	b := newBrowser(t)
	status, body := b.open(link)
	if status != http.StatusOK {
		t.Fatalf("unexpected status %d", status)
	}
	for _, want := range []string{SessionFilesSource, ">bitwarden<", "Vault prod", "c8y-folder", "op is not signed in"} {
		if !strings.Contains(body, want) {
			t.Errorf("page should contain %q", want)
		}
	}

	status, body = b.post("/select", url.Values{"csrf": {b.csrf}, "session": {"1"}}, nil)
	if status != http.StatusOK || !strings.Contains(body, "Session selected") {
		t.Fatalf("unexpected response %d: %s", status, body)
	}
	r := waitPicker(t, results)
	if r.err != nil || r.selection.Choice.URI != "bitwarden://item-1" || r.selection.Choice.Provider != "bitwarden" {
		t.Errorf("unexpected result: %+v", r)
	}
}

func TestWebPickerOnlyProblems(t *testing.T) {
	_, err := (&WebPicker{}).Pick(context.Background(), "Select", nil, SourceError{Source: "bitwarden", Message: "BW_SESSION is not set"})
	if err == nil || !strings.Contains(err.Error(), "bitwarden: BW_SESSION is not set") {
		t.Errorf("expected the provider error, got %v", err)
	}
}

func TestShowGroupHeadings(t *testing.T) {
	files := SessionChoice{Name: "dev", Path: "/s/dev.json"}
	provider := SessionChoice{Name: "vault", Source: "bitwarden", Provider: "bitwarden", URI: "bitwarden://1"}
	cases := []struct {
		name     string
		choices  []SessionChoice
		problems []SourceError
		want     bool
	}{
		{name: "only session files", choices: []SessionChoice{files}, want: false},
		{name: "only one provider", choices: []SessionChoice{provider}, want: true},
		{name: "files and provider", choices: []SessionChoice{files, provider}, want: true},
		{name: "files and failed provider", choices: []SessionChoice{files}, problems: []SourceError{{Source: "bitwarden", Message: "locked"}}, want: true},
	}
	for _, tc := range cases {
		if got := showGroupHeadings(groupChoices(tc.choices, tc.problems)); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}
