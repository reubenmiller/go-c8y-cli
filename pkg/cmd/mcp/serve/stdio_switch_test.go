package serve

import (
	"context"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mcpserver"
	"github.com/spf13/viper"
)

func newSwitchTestSessions(ceiling config.SessionMode) *stdioSessions {
	cfg := config.NewConfig(viper.New())
	policy := mcpserver.DefaultPolicy()
	policy.Mode = ceiling
	return &stdioSessions{
		server: &mcpserver.Server{Policy: policy},
		cfg:    cfg,
		log:    cfg.Logger,
		sources: []providerSource{
			{Ref: "bitwarden"},
		},
		listProviderSessions: func(ctx context.Context, ref string, args []string) (string, []extensions.ProviderSession, error) {
			return "bitwarden", []extensions.ProviderSession{
				{SessionURI: "bitwarden://1", Name: "Dev", Host: "https://dev.example.com"},
				{SessionURI: "bitwarden://2", Name: "Dev", Host: "https://other.example.com"},
				{SessionURI: "bitwarden://3", Name: "Prod", Host: "https://prod.example.com"},
			}, nil
		},
	}
}

func TestSwitchToFindChoice(t *testing.T) {
	s := newSwitchTestSessions(config.SessionModeDev)
	ctx := context.Background()

	choice, err := s.findChoice(ctx, "prod")
	if err != nil || choice.URI != "bitwarden://3" {
		t.Fatalf("expected the Prod session: %+v, %v", choice, err)
	}
	if choice, err = s.findChoice(ctx, "bitwarden://1"); err != nil || choice.URI != "bitwarden://1" {
		t.Fatalf("expected a match by uri: %+v, %v", choice, err)
	}
	if _, err = s.findChoice(ctx, "dev"); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected an ambiguity error: %v", err)
	}
	if _, err = s.findChoice(ctx, "missing"); err == nil || !strings.Contains(err.Error(), "unknown session") {
		t.Fatalf("expected an unknown session error: %v", err)
	}
}

func TestSessionNamesUsesListedProviderSessions(t *testing.T) {
	s := newSwitchTestSessions(config.SessionModeDev)
	if names := s.sessionNames(context.Background()); len(names) != 0 {
		t.Fatalf("providers should not be listed for a completion: %v", names)
	}
	s.providerChoices(context.Background())
	names := s.sessionNames(context.Background())
	if len(names) < 2 {
		t.Fatalf("expected the listed provider sessions: %v", names)
	}
}

func TestSwitchToModeOnly(t *testing.T) {
	s := newSwitchTestSessions(config.SessionModeQual)
	if _, err := s.switchTo(context.Background(), "", "qual"); err == nil {
		t.Fatal("expected an error without a selected session")
	}
	s.server.SetSessionKey("k", mcpserver.SessionInfo{SessionName: "k", SessionMode: "prod"}, &mcpserver.Runner{Mode: config.SessionModeProduction})
	message, err := s.switchTo(context.Background(), "", "qual")
	if err != nil || !strings.HasPrefix(message, "Mode is now qual (override; the session's own mode is prod)") {
		t.Fatalf("unexpected result %q, %v", message, err)
	}
	if _, err := s.switchTo(context.Background(), "", "dev"); err == nil {
		t.Fatal("expected a mode above the ceiling to be rejected")
	}
	if _, err := s.switchTo(context.Background(), "", "bogus"); err == nil {
		t.Fatal("expected an invalid mode to be rejected")
	}
}

func TestApplyPickedMode(t *testing.T) {
	s := newSwitchTestSessions(config.SessionModeQual)
	s.server.SetSessionKey("k", mcpserver.SessionInfo{SessionName: "k", SessionMode: "prod"}, &mcpserver.Runner{Mode: config.SessionModeProduction})
	choice := mcpserver.SessionChoice{SessionMode: "prod"}

	if err := s.applyPickedMode(&mcpserver.Selection{Choice: choice, Mode: config.SessionModeQual}); err != nil {
		t.Fatal(err)
	}
	if _, runner := s.server.Active(); runner.Mode != config.SessionModeQual {
		t.Fatalf("expected the override to be applied, got %s", runner.Mode)
	}
	// the session's own mode resets the override
	if err := s.applyPickedMode(&mcpserver.Selection{Choice: choice, Mode: config.SessionModeProduction, ModeOnly: true}); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.server.ModeOverride("k"); ok {
		t.Fatal("the override should have been reset")
	}
	// no mode chosen: nothing changes
	if err := s.applyPickedMode(&mcpserver.Selection{Choice: choice}); err != nil {
		t.Fatal(err)
	}
	// the ceiling is never exceeded
	if err := s.applyPickedMode(&mcpserver.Selection{Choice: choice, Mode: config.SessionModeDev}); err == nil {
		t.Fatal("expected a mode above the ceiling to be rejected")
	}
}

func TestChoiceModes(t *testing.T) {
	s := newSwitchTestSessions(config.SessionModeQual)
	_, sessions, _ := s.listProviderSessions(context.Background(), "bitwarden", nil)
	if len(sessions) == 0 {
		t.Fatal("expected sessions")
	}
	choices := s.providerChoices(context.Background())[0].choices
	if choices[0].MaxMode != "qual" || choices[0].SessionMode != "prod" || choices[0].Mode != "prod" {
		t.Fatalf("unexpected modes: %+v", choices[0])
	}
}
