package extensions

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func parseManifest(t *testing.T, contents string) *Manifest {
	t.Helper()
	m := &Manifest{}
	if err := yaml.Unmarshal([]byte(contents), m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestSessionProvidersParsing(t *testing.T) {
	m := parseManifest(t, `
binary: {}
sessionProviders:
  - name: bitwarden
    description: Select a session from Bitwarden
    exec: [list]
    secrets: [BW_SESSION]
    format: json
  - name: env-file
    format: AUTO
`)
	providers, warnings := m.ValidSessionProviders()
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}
	p := providers[0]
	if p.Name != "bitwarden" || p.Description == "" || strings.Join(p.Exec, " ") != "list" || strings.Join(p.Secrets, ",") != "BW_SESSION" {
		t.Errorf("unexpected provider: %+v", p)
	}
	if p.OutputFormat() != "json" {
		t.Errorf("unexpected format: %s", p.OutputFormat())
	}
	if providers[1].OutputFormat() != "" {
		t.Errorf("auto format should be empty, got %s", providers[1].OutputFormat())
	}
}

func TestSessionProvidersWithoutBinary(t *testing.T) {
	m := parseManifest(t, "sessionProviders:\n  - name: bitwarden\n")
	providers, warnings := m.ValidSessionProviders()
	if len(providers) != 0 {
		t.Errorf("providers should be ignored without a binary: %v", providers)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "binary") {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestSessionProvidersValidation(t *testing.T) {
	m := parseManifest(t, `
binary: {}
sessionProviders:
  - name: vault
    exec: [first]
  - name: vault
    exec: [second]
  - name: Invalid_Name
  - name: ""
  - name: other
    format: yaml
`)
	providers, warnings := m.ValidSessionProviders()
	if len(providers) != 1 || providers[0].Name != "vault" || providers[0].Exec[0] != "first" {
		t.Errorf("unexpected providers: %+v", providers)
	}
	if len(warnings) != 4 {
		t.Fatalf("expected 4 warnings, got %v", warnings)
	}
	for i, want := range []string{"duplicate", "invalid", "invalid", "format"} {
		if !strings.Contains(warnings[i], want) {
			t.Errorf("warning %d should contain %q: %s", i, want, warnings[i])
		}
	}
}

func TestNoSessionProviders(t *testing.T) {
	providers, warnings := parseManifest(t, "version: v1\n").ValidSessionProviders()
	if providers != nil || warnings != nil {
		t.Errorf("expected no providers and warnings: %v %v", providers, warnings)
	}
}
