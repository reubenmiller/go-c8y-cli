package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestGetDeniedPathsFromEnv(t *testing.T) {
	t.Setenv("C8Y_SETTINGS_FILES_DENIEDPATHS", strings.Join([]string{"/tmp/a", "", "~/b"}, string(os.PathListSeparator)))
	c := NewConfig(viper.New())
	got := c.GetDeniedPaths()
	if len(got) != 2 || got[0] != "/tmp/a" || strings.HasPrefix(got[1], "~") || !strings.HasSuffix(got[1], "b") {
		t.Fatalf("unexpected denied paths: %v", got)
	}
}

func TestGetDeniedPathsFromSettingsFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(file, []byte(`{"settings": {"files": {"deniedPaths": ["/tmp/a", "/tmp/b"]}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	v := viper.New()
	c := NewConfig(v)
	v.SetConfigFile(file)
	if err := v.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	if got := c.GetDeniedPaths(); len(got) != 2 || got[0] != "/tmp/a" || got[1] != "/tmp/b" {
		t.Fatalf("unexpected denied paths: %v", got)
	}

	// the environment variable takes precedence
	t.Setenv("C8Y_SETTINGS_FILES_DENIEDPATHS", "/tmp/c")
	if got := c.GetDeniedPaths(); len(got) != 1 || got[0] != "/tmp/c" {
		t.Fatalf("expected the env variable to override the settings file: %v", got)
	}
}

func TestGetDeniedPathsUnset(t *testing.T) {
	t.Setenv("C8Y_SETTINGS_FILES_DENIEDPATHS", "")
	if got := NewConfig(viper.New()).GetDeniedPaths(); len(got) != 0 {
		t.Fatalf("expected no denied paths, got %v", got)
	}
}
