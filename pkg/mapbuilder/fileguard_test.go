package mapbuilder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileguard"
)

func TestJsonnetImportOfDeniedFile(t *testing.T) {
	root := t.TempDir()
	sessions := filepath.Join(root, "sessions")
	templates := filepath.Join(root, "templates")
	for _, d := range []string{sessions, templates} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	session := filepath.Join(sessions, "dev.json")
	allowed := filepath.Join(templates, "base.json")
	if err := os.WriteFile(session, []byte(`{"password":"secret-value"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(allowed, []byte(`{"type":"c8y_Device"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	fileguard.SetDeniedPaths(sessions)
	t.Cleanup(func() { fileguard.SetDeniedPaths() })

	for _, snippet := range []string{
		"{leak: import '" + session + "'}",
		"{leak: importstr '" + session + "'}",
		"{leak: importbin '" + session + "'}",
	} {
		out, err := evaluateJsonnet("", snippet)
		if err == nil || strings.Contains(out, "secret-value") {
			t.Errorf("expected the import to be denied: %s => %s, %v", snippet, out, err)
		}
		if err != nil && strings.Contains(err.Error(), "secret-value") {
			t.Errorf("error should not contain the file contents: %v", err)
		}
	}

	out, err := evaluateJsonnet("", "{base: import '"+allowed+"'}")
	if err != nil || !strings.Contains(out, "c8y_Device") {
		t.Errorf("expected the import of an allowed file to work: %s, %v", out, err)
	}
}
