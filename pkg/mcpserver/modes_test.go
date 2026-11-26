package mcpserver

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

func TestResolveMode(t *testing.T) {
	const (
		prod  = config.SessionModeProduction
		qual  = config.SessionModeQual
		dev   = config.SessionModeDev
		unset = config.SessionModeUnset
	)
	cases := []struct{ ceiling, session, override, expected config.SessionMode }{
		{unset, unset, unset, prod},
		{unset, qual, unset, qual},
		{unset, prod, dev, dev},
		{qual, prod, dev, qual},
		{prod, dev, unset, prod},
		{dev, dev, prod, prod},
		{dev, prod, config.SessionModeCI, dev},
		{unset, config.SessionModeCI, unset, dev},
	}
	for _, tc := range cases {
		if got := ResolveMode(tc.ceiling, tc.session, tc.override); got != tc.expected {
			t.Errorf("ResolveMode(%s, %s, %s) = %s, expected %s", tc.ceiling, tc.session, tc.override, got, tc.expected)
		}
	}
}

func newModeServer(ceiling config.SessionMode, logs *[]string) *Server {
	s := &Server{Policy: Policy{Mode: ceiling}}
	s.Logf = func(format string, args ...any) {
		*logs = append(*logs, strings.TrimSpace(fmt.Sprintf(format, args...)))
	}
	return s
}

func TestServerSetMode(t *testing.T) {
	var logs []string
	s := newModeServer(config.SessionModeQual, &logs)
	runnerA := &Runner{Mode: config.SessionModeProduction, Auth: &InteractiveLogin{}}
	s.SetSessionKey("/a", SessionInfo{SessionName: "a", SessionMode: "prod"}, runnerA)

	if _, err := s.SetMode(config.SessionModeDev, "test"); err == nil {
		t.Fatal("expected a mode above the ceiling to be rejected")
	}
	if _, runner := s.Active(); runner.Mode != config.SessionModeProduction {
		t.Fatalf("mode changed after a rejected change: %s", runner.Mode)
	}

	change, err := s.SetMode(config.SessionModeQual, "picker")
	if err != nil {
		t.Fatal(err)
	}
	if change.From != config.SessionModeProduction || change.To != config.SessionModeQual || change.Source != ModeSourceOverride {
		t.Errorf("unexpected change %+v", change)
	}
	_, runner := s.Active()
	if runner.Mode != config.SessionModeQual || runner.Auth != runnerA.Auth {
		t.Errorf("runner should use qual with the same auth, got %s", runner.Mode)
	}
	if runnerA.Mode != config.SessionModeProduction {
		t.Error("the original runner must not be modified")
	}
	if len(logs) != 1 || logs[0] != "mode override: prod -> qual (session a, via picker)" {
		t.Errorf("unexpected log %v", logs)
	}

	// switching to another session and back keeps the override
	s.SetSessionKey("/b", SessionInfo{SessionName: "b", SessionMode: "prod"}, &Runner{Mode: config.SessionModeProduction})
	if _, runner := s.Active(); runner.Mode != config.SessionModeProduction {
		t.Errorf("other session must keep its own mode, got %s", runner.Mode)
	}
	s.SetSessionKey("/a", SessionInfo{SessionName: "a", SessionMode: "prod"}, runnerA)
	if _, runner := s.Active(); runner.Mode != config.SessionModeQual {
		t.Errorf("override should be kept, got %s", runner.Mode)
	}

	change, err = s.SetMode(config.SessionModeUnset, "reset")
	if err != nil {
		t.Fatal(err)
	}
	if change.To != config.SessionModeProduction || change.Source != ModeSourceSession {
		t.Errorf("unexpected reset %+v", change)
	}
	if _, ok := s.ModeOverride("/a"); ok {
		t.Error("override should be removed")
	}
}

func TestServerSetModeWithoutSession(t *testing.T) {
	var logs []string
	s := newModeServer(config.SessionModeUnset, &logs)
	if _, err := s.SetMode(config.SessionModeQual, "test"); err == nil {
		t.Fatal("expected an error without a session")
	}
}

func TestCheckMode(t *testing.T) {
	if err := CheckMode(config.SessionModeProduction, config.SessionModeProduction); err != nil {
		t.Error(err)
	}
	if err := CheckMode(config.SessionModeProduction, config.SessionModeQual); err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Errorf("unexpected %v", err)
	}
	if err := CheckMode(config.SessionModeQual, config.SessionModeDev); err == nil {
		t.Error("expected error")
	}
	if err := CheckMode(config.SessionModeDev, config.SessionModeQual); err != nil {
		t.Error(err)
	}
}

func TestSessionToolShowsModeSource(t *testing.T) {
	var logs []string
	s := newModeServer(config.SessionModeQual, &logs)
	s.SetSessionKey("/a", SessionInfo{SessionName: "a", SessionMode: "prod"}, &Runner{Mode: config.SessionModeProduction})
	show := func() string {
		res, _, err := s.handleSession(context.Background(), nil, struct{}{})
		if err != nil {
			t.Fatal(err)
		}
		return res.Content[0].(*mcp.TextContent).Text
	}
	out := show()
	for _, want := range []string{`"modeSource": "session"`, `"sessionMode": "prod"`, `"maxMode": "qual"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
	if _, err := s.SetMode(config.SessionModeQual, "test"); err != nil {
		t.Fatal(err)
	}
	if out := show(); !strings.Contains(out, `"modeSource": "override"`) || !strings.Contains(out, `"mode": "qual"`) {
		t.Errorf("unexpected %s", out)
	}
}
