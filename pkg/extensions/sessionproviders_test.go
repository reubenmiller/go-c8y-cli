package extensions

import (
	"errors"
	"sort"
	"strings"
	"testing"
)

func mockExtension(name string, manifest *Manifest) *ExtensionMock {
	return &ExtensionMock{
		NameFunc:     func() string { return name },
		ManifestFunc: func() (*Manifest, error) { return manifest, nil },
	}
}

func providerManifest(names ...string) *Manifest {
	m := &Manifest{Binary: &BinarySpec{}}
	for _, name := range names {
		m.SessionProviders = append(m.SessionProviders, SessionProvider{Name: name, Description: name + " provider"})
	}
	return m
}

func testProviders(t *testing.T) []InstalledSessionProvider {
	t.Helper()
	var warnings []string
	providers := ListSessionProviders([]Extension{
		mockExtension("session-bitwarden", providerManifest("bitwarden", "vault")),
		mockExtension("session-hashicorp", providerManifest("vault")),
		mockExtension("nobinary", &Manifest{SessionProviders: []SessionProvider{{Name: "ignored"}}}),
		mockExtension("other", &Manifest{}),
	}, func(ext Extension, msg string) {
		warnings = append(warnings, ext.Name()+": "+msg)
	})
	if len(warnings) != 1 || !strings.HasPrefix(warnings[0], "nobinary:") {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	return providers
}

func TestListSessionProviders(t *testing.T) {
	providers := testProviders(t)
	got := strings.Join(providerRefs(providers), ",")
	if got != "session-bitwarden/bitwarden,session-bitwarden/vault,session-hashicorp/vault" {
		t.Errorf("unexpected providers: %s", got)
	}
}

func TestResolveSessionProvider(t *testing.T) {
	providers := testProviders(t)
	cases := []struct {
		ref     string
		want    string
		wantErr error
	}{
		{"bitwarden", "session-bitwarden/bitwarden", nil},
		{"session-bitwarden/bitwarden", "session-bitwarden/bitwarden", nil},
		{"c8y-session-bitwarden/bitwarden", "session-bitwarden/bitwarden", nil},
		{"session-hashicorp/vault", "session-hashicorp/vault", nil},
		{"vault", "", ErrAmbiguousSessionProvider},
		{"missing", "", ErrUnknownSessionProvider},
		{"session-bitwarden/missing", "", ErrUnknownSessionProvider},
		{"ignored", "", ErrUnknownSessionProvider},
	}
	for _, tc := range cases {
		p, err := ResolveSessionProvider(providers, tc.ref)
		if tc.wantErr != nil {
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("%s: expected %v, got %v", tc.ref, tc.wantErr, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected error: %s", tc.ref, err)
			continue
		}
		if p.Ref() != tc.want {
			t.Errorf("%s: expected %s, got %s", tc.ref, tc.want, p.Ref())
		}
	}

	_, err := ResolveSessionProvider(providers, "vault")
	if !strings.Contains(err.Error(), "session-bitwarden/vault") || !strings.Contains(err.Error(), "session-hashicorp/vault") {
		t.Errorf("ambiguous error should list the qualified references: %s", err)
	}
	_, err = ResolveSessionProvider(providers, "missing")
	if !strings.Contains(err.Error(), "session-bitwarden/bitwarden") {
		t.Errorf("unknown error should list the available providers: %s", err)
	}
	_, err = ResolveSessionProvider(nil, "missing")
	if !strings.Contains(err.Error(), "<none>") {
		t.Errorf("unexpected error without providers: %s", err)
	}
}

func TestSessionProviderCompletions(t *testing.T) {
	got := SessionProviderCompletions(testProviders(t))
	want := []string{
		"bitwarden\tbitwarden provider",
		"session-bitwarden/vault\tvault provider",
		"session-hashicorp/vault\tvault provider",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("unexpected completions: %q", got)
	}
}

func TestMergeEnv(t *testing.T) {
	base := []string{"PATH=/bin", "C8Y_HOST=https://old", "C8Y_TOKEN=old", "C8Y_PASSWORD=keep", "OTHER=1"}
	env := MergeEnv(base,
		map[string]string{"C8Y_HOST": "https://new", "C8Y_EXTENSION_NAME": "foo"},
		map[string]bool{"C8Y_HOST": true, "C8Y_TOKEN": true},
	)
	sort.Strings(env)
	want := []string{"C8Y_EXTENSION_NAME=foo", "C8Y_HOST=https://new", "C8Y_PASSWORD=keep", "OTHER=1", "PATH=/bin"}
	if strings.Join(env, "|") != strings.Join(want, "|") {
		t.Errorf("unexpected env: %v", env)
	}
}

func TestVariables(t *testing.T) {
	ext := &ExtensionMock{
		NameFunc: func() string { return "foo" },
		PathFunc: func() string { return "/data/c8y-foo" },
	}
	vars := Variables(ext, "c8y")
	if vars[EnvExtensionName] != "foo" || vars[EnvExtensionDir] != "/data/c8y-foo" || vars[EnvContextVersion] != ContextVersion || vars[EnvBin] == "" {
		t.Errorf("unexpected variables: %v", vars)
	}
}

func TestValidSessionProvidersListAndSelect(t *testing.T) {
	m := &Manifest{Binary: &BinarySpec{}, SessionProviders: []SessionProvider{
		{Name: "both", List: []string{"list"}, Select: []string{"get"}},
		{Name: "listonly", List: []string{"list"}},
		{Name: "none"},
	}}
	providers, warnings := m.ValidSessionProviders()
	if len(providers) != 3 {
		t.Fatalf("all providers should still be usable: %+v", providers)
	}
	if !providers[0].CanList() || providers[1].CanList() || providers[2].CanList() {
		t.Errorf("unexpected CanList: %+v", providers)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "listonly") {
		t.Errorf("expected a warning for the provider with only list: %v", warnings)
	}
}

func TestParseProviderSessions(t *testing.T) {
	sessions, warnings, err := ParseProviderSessions([]byte(`[{"sessionUri":"a://1","host":"https://a","token":"x","password":"y"},{"host":"https://b"}]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].SessionURI != "a://1" || len(warnings) != 1 {
		t.Errorf("unexpected result: %+v %v", sessions, warnings)
	}
	if _, _, err := ParseProviderSessions([]byte(`{"host":"https://a"}`)); err == nil {
		t.Errorf("an object should not be accepted")
	}
}
