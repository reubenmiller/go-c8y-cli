package serve

import (
	"reflect"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/spf13/viper"
)

func testServe(t *testing.T, v *viper.Viper, flags ...string) (*CmdServe, *config.Config) {
	t.Helper()
	if v == nil {
		v = viper.New()
	}
	cfg := config.NewConfig(v)
	manifest := &extensions.Manifest{Binary: &extensions.BinarySpec{}, SessionProviders: []extensions.SessionProvider{
		{Name: "bitwarden", Exec: []string{"list"}, List: []string{"list", "--json"}, Select: []string{"get"}},
		{Name: "interactive", Exec: []string{"pick"}},
		{Name: "file", List: []string{"list"}, Select: []string{"get"}},
	}}
	ext := &extensions.ExtensionMock{
		NameFunc:     func() string { return "session-bitwarden" },
		ManifestFunc: func() (*extensions.Manifest, error) { return manifest, nil },
	}
	f := &cmdutil.Factory{
		Logger: func() (*logger.Logger, error) { return cfg.Logger, nil },
		ExtensionManager: func() extensions.ExtensionManager {
			return &extensions.ExtensionManagerMock{ListFunc: func() []extensions.Extension { return []extensions.Extension{ext} }}
		},
	}
	return &CmdServe{factory: f, sessionProviders: flags}, cfg
}

func TestProviderSources(t *testing.T) {
	cases := []struct {
		name  string
		flags []string
		want  []providerSource
	}{
		{name: "default", want: []providerSource{{Ref: "file"}}},
		{name: "provider only", flags: []string{"bitwarden"}, want: []providerSource{{Ref: "bitwarden", Args: []string{}}}},
		{name: "provider first", flags: []string{"bitwarden --folder 'my c8y'", "file"}, want: []providerSource{{Ref: "bitwarden", Args: []string{"--folder", "my c8y"}}, {Ref: "file"}}},
		{name: "file only", flags: []string{"file"}, want: []providerSource{{Ref: "file"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n, cfg := testServe(t, nil, tc.flags...)
			got, err := n.providerSources(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestProviderSourcesFromConfig(t *testing.T) {
	v := viper.New()
	v.Set(config.SettingsSessionProviderType, config.ProviderTypeExtension)
	v.Set(config.SettingsSessionProviderExtension, "bitwarden")
	n, cfg := testServe(t, v)
	got, err := n.providerSources(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Ref != "file" || got[1].Ref != "bitwarden" {
		t.Errorf("expected the session files and the configured provider: %+v", got)
	}

	// a configured provider which can't list is skipped (it still works for sessions login)
	v.Set(config.SettingsSessionProviderExtension, "interactive")
	got, err = n.providerSources(cfg)
	if err != nil || len(got) != 1 || got[0].Ref != "file" {
		t.Errorf("expected only the session files: %+v %v", got, err)
	}

	// the flag replaces the configured sources
	n.sessionProviders = []string{"bitwarden"}
	v.Set(config.SettingsSessionProviderExtension, "bitwarden")
	got, err = n.providerSources(cfg)
	if err != nil || len(got) != 1 || got[0].Ref != "bitwarden" {
		t.Errorf("expected only the flag's provider: %+v %v", got, err)
	}
}

func TestProviderSourcesErrors(t *testing.T) {
	cases := map[string][]string{
		"does not accept arguments": {"file --folder x"},
		"more than once":            {"file", "file"},
		"can't list its sessions":   {"interactive"},
		"unknown session provider":  {"keepass"},
		// the extension's provider named "file" is ignored, as the name is reserved
		`"session-bitwarden/file"`: {"session-bitwarden/file"},
	}
	for want, flags := range cases {
		n, cfg := testServe(t, nil, flags...)
		_, err := n.providerSources(cfg)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("flags %v: expected an error containing %q, got %v", flags, want, err)
		}
	}
}
