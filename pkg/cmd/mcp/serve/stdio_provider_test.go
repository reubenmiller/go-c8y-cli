package serve

import (
	"context"
	"errors"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mcpserver"
	"github.com/spf13/viper"
)

func TestProviderChoices(t *testing.T) {
	cfg := config.NewConfig(viper.New())
	policy := mcpserver.DefaultPolicy()
	policy.Mode = config.SessionModeQual
	s := &stdioSessions{
		server: &mcpserver.Server{Policy: policy},
		cfg:    cfg,
		log:    cfg.Logger,
		sources: []providerSource{
			{Ref: FileSessionSource},
			{Ref: "bitwarden", Args: []string{"--folder", "c8y"}},
			{Ref: "session-op/onepassword"},
		},
		listProviderSessions: func(ctx context.Context, ref string, args []string) (string, []extensions.ProviderSession, error) {
			if ref == "bitwarden" {
				if len(args) != 2 || args[1] != "c8y" {
					t.Errorf("list args should be passed: %v", args)
				}
				return "bitwarden", []extensions.ProviderSession{
					{SessionURI: "bitwarden://1", Name: "Dev", Host: "https://dev.example.com/", Mode: "dev", Group: "c8y"},
					{SessionURI: "bitwarden://2", Host: "https://prod.example.com"},
				}, nil
			}
			return "onepassword", nil, errors.New("op is not signed in")
		},
	}
	s.current = providerKey("bitwarden", "bitwarden://2")

	listed := s.providerChoices(context.Background())
	if len(listed) != 3 || len(listed[0].choices) != 0 || len(listed[0].problems) != 0 {
		t.Fatalf("the result should have one item per source, and none for the session files: %+v", listed)
	}
	choices := listed[1].choices
	problems := listed[2].problems
	if len(choices) != 2 || len(problems) != 1 || len(listed[1].problems) != 0 || len(listed[2].choices) != 0 {
		t.Fatalf("unexpected result: %+v", listed)
	}
	dev := choices[0]
	if dev.Source != "bitwarden" || dev.Provider != "bitwarden" || dev.URI != "bitwarden://1" || dev.Host != "https://dev.example.com" || dev.Group != "c8y" {
		t.Errorf("unexpected choice: %+v", dev)
	}
	if dev.Mode != "qual" {
		t.Errorf("the session mode should be capped by the policy: %s", dev.Mode)
	}
	if choices[1].Mode != "prod" || !choices[1].Current || choices[1].Name != "bitwarden://2" {
		t.Errorf("unexpected choice: %+v", choices[1])
	}
	if problems[0].Source != "onepassword" || problems[0].Message != "op is not signed in" {
		t.Errorf("unexpected problem: %+v", problems[0])
	}

	session, login, err := s.newProviderSession(dev)
	if err != nil {
		t.Fatal(err)
	}
	if login.Provider != "bitwarden" || login.SessionURI != "bitwarden://1" || login.SessionFile != "" {
		t.Errorf("unexpected login: %+v", login)
	}
	if session.info.Provider != "bitwarden" || session.info.SessionName != "Dev" || session.runner.Mode != config.SessionModeQual || session.runner.SessionFile != "" {
		t.Errorf("unexpected session: %+v %+v", session.info, session.runner)
	}

	if _, _, err := s.newProviderSession(mcpserver.SessionChoice{Provider: "bitwarden", URI: "bitwarden://unknown"}); err == nil {
		t.Errorf("unlisted sessions should be rejected")
	}
}

func TestChoicesFollowSourceOrder(t *testing.T) {
	v := viper.New()
	v.Set("settings.session.home", t.TempDir())
	cfg := config.NewConfig(v)
	newSessions := func(sources ...providerSource) *stdioSessions {
		return &stdioSessions{
			server:  &mcpserver.Server{Policy: mcpserver.DefaultPolicy()},
			cfg:     cfg,
			log:     cfg.Logger,
			sources: sources,
			listProviderSessions: func(ctx context.Context, ref string, args []string) (string, []extensions.ProviderSession, error) {
				return ref, []extensions.ProviderSession{{SessionURI: ref + "://1", Host: "https://" + ref + ".example.com"}}, nil
			},
		}
	}

	choices, problems, err := newSessions(providerSource{Ref: "vault-b"}, providerSource{Ref: "vault-a"}).choices(context.Background())
	if err != nil || len(problems) != 0 {
		t.Fatal(err, problems)
	}
	if len(choices) != 2 || choices[0].Provider != "vault-b" || choices[1].Provider != "vault-a" {
		t.Errorf("the sessions should be in the order of the sources: %+v", choices)
	}
	for _, c := range choices {
		if c.Path != "" {
			t.Errorf("session files should not be listed without the file source: %+v", c)
		}
	}
}
