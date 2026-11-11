package flags

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileguard"
)

func TestResolveContentsOfDeniedFile(t *testing.T) {
	root := t.TempDir()
	sessions := filepath.Join(root, "sessions")
	if err := os.MkdirAll(sessions, 0o700); err != nil {
		t.Fatal(err)
	}
	session := filepath.Join(sessions, "dev.json")
	allowed := filepath.Join(root, "data.json")
	for _, f := range []string{session, allowed} {
		if err := os.WriteFile(f, []byte(`{"password":"secret-value"}`), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	fileguard.SetDeniedPaths(sessions)
	t.Cleanup(func() { fileguard.SetDeniedPaths() })

	if v, err := resolveContents(session); !errors.Is(err, fileguard.ErrDenied) || v != "" {
		t.Errorf("expected a denied error, got %q, %v", v, err)
	}
	if v, err := getContents(session); !errors.Is(err, fileguard.ErrDenied) || v != "" {
		t.Errorf("expected a denied error for templates, got %q, %v", v, err)
	}
	if v, err := resolveContents(allowed); err != nil || v == allowed {
		t.Errorf("expected the contents of the allowed file, got %q, %v", v, err)
	}
	if v, err := resolveContents("name=value"); err != nil || v != "name=value" {
		t.Errorf("expected a non-file value to be returned as is, got %q, %v", v, err)
	}
}
