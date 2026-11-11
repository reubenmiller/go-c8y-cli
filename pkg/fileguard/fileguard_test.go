package fileguard

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	root := t.TempDir()
	sessions := filepath.Join(root, "sessions")
	other := filepath.Join(root, "work")
	for _, d := range []string{sessions, other} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	secret := filepath.Join(sessions, "dev.json")
	if err := os.WriteFile(secret, []byte(`{"password":"x"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	// symlink from an allowed location into the denied folder
	link := filepath.Join(other, "link.json")
	linkErr := os.Symlink(secret, link)

	denied := parse(sessions)
	cases := []struct {
		name   string
		path   string
		denied bool
	}{
		{"file in denied folder", secret, true},
		{"denied folder itself", sessions, true},
		{"new file in denied folder", filepath.Join(sessions, "new.json"), true},
		{"dot dot path", filepath.Join(other, "..", "sessions", "dev.json"), true},
		{"allowed file", filepath.Join(other, "template.jsonnet"), false},
		{"similar prefix", sessions + "-backup/dev.json", false},
	}
	if linkErr == nil {
		cases = append(cases, struct {
			name   string
			path   string
			denied bool
		}{"symlink into denied folder", link, true})
	}
	if caseInsensitive {
		cases = append(cases, struct {
			name   string
			path   string
			denied bool
		}{"different case", strings.ToUpper(secret[:len(secret)-len("dev.json")]) + "dev.json", true})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := check(tc.path, denied)
			if tc.denied != (err != nil) {
				t.Fatalf("check(%s) = %v, want denied=%v", tc.path, err, tc.denied)
			}
			if err != nil && !errors.Is(err, ErrDenied) {
				t.Fatalf("unexpected error type: %v", err)
			}
		})
	}
}

func TestCheckRelativePath(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "sessions"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	if err := check("sessions/dev.json", parse(filepath.Join(root, "sessions"))); err == nil {
		t.Fatal("expected a relative path into the denied folder to be denied")
	}
}

func TestSetDeniedPaths(t *testing.T) {
	root := t.TempDir()
	t.Cleanup(func() { SetDeniedPaths() })
	SetDeniedPaths(filepath.Join(root, "a"), filepath.Join(root, "b"))
	for _, p := range []string{"a/x.json", "b/y.json"} {
		if err := Check(filepath.Join(root, p)); !errors.Is(err, ErrDenied) {
			t.Errorf("expected %s to be denied, got %v", p, err)
		}
	}
	if err := Check(filepath.Join(root, "c.json")); err != nil {
		t.Errorf("expected an allowed file, got %v", err)
	}
	SetDeniedPaths()
	if Enabled() {
		t.Error("expected no denied paths after clearing them")
	}
}

func TestParseIgnoresEmptyEntries(t *testing.T) {
	if got := parse("", " "); len(got) != 0 {
		t.Fatalf("expected no denied paths, got %v", got)
	}
}
