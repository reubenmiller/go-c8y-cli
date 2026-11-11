package flags

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileguard"
	"github.com/spf13/cobra"
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

// A data flag of type string (instead of a string array) must not return the error
// from the string array lookup, e.g. "bulkoperations create --operation"
func TestDataValueFromStringFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("operation", "", "")
	if err := cmd.Flags().Set("operation", "c8y_Restart={}"); err != nil {
		t.Fatal(err)
	}
	dst, value, err := WithDataValueAdvanced(false, false, "operation", "operationPrototype")(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, ok := value.(map[string]interface{})
	if dst != "operationPrototype" || !ok || data["c8y_Restart"] == nil {
		t.Fatalf("unexpected value: dst=%s, value=%v", dst, value)
	}
}
