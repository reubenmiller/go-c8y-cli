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
