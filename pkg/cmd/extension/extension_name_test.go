package extension

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
)

func TestReadDeclaredName(t *testing.T) {
	dir := t.TempDir()

	name, err := readDeclaredName(dir)
	if err != nil || name != "" {
		t.Fatalf("missing file: got %q, %v", name, err)
	}

	write := func(content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, "extension.yaml"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	write("version: v1\n")
	if name, err := readDeclaredName(dir); err != nil || name != "" {
		t.Fatalf("no name: got %q, %v", name, err)
	}

	write("version: v1\nname: Analytics\n")
	if name, err := readDeclaredName(dir); err != nil || name != "Analytics" {
		t.Fatalf("name: got %q, %v", name, err)
	}

	write("name: [\n")
	if _, err := readDeclaredName(dir); err == nil {
		t.Fatal("invalid yaml: expected an error")
	}
}

func TestResolveInstallFolder(t *testing.T) {
	withName := t.TempDir()
	if err := os.WriteFile(filepath.Join(withName, "extension.yaml"), []byte("name: Analytics\n"), 0644); err != nil {
		t.Fatal(err)
	}
	withoutName := t.TempDir()

	tests := []struct {
		name          string
		flag          string
		dir           string
		fallback      string
		requirePrefix bool
		want          string
		wantErr       bool
	}{
		{"flag wins over the declared name", "custom", withName, "c8y-repo", true, "c8y-custom", false},
		{"flag with prefix", "c8y-custom", withName, "c8y-repo", true, "c8y-custom", false},
		{"declared name wins over the repository name", "", withName, "c8y-repo", true, "c8y-analytics", false},
		{"declared name without repository prefix", "", withName, "repo", true, "c8y-analytics", false},
		{"repository name", "", withoutName, "c8y-repo", true, "c8y-repo", false},
		{"repository name containing the prefix", "", withoutName, "go-c8y-repo", true, "go-c8y-repo", false},
		{"repository name without the prefix", "", withoutName, "repo", true, "", true},
		{"local folder without the prefix", "", withoutName, "repo", false, "c8y-repo", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveInstallFolder(extensions.InstallOptions{Name: tt.flag}, tt.dir, tt.fallback, tt.requirePrefix)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
