package powershell

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestGoldenSpecOutput verifies that the generator reproduces every committed
// tools/PSc8y/Public/*.ps1 cmdlet and tools/PSc8y/Tests/*.auto.Tests.ps1
// test byte-for-byte from the api/spec/json specifications.
func TestGoldenSpecOutput(t *testing.T) {
	repoRoot := "../../.."
	specDir := filepath.Join(repoRoot, "api", "spec", "json")
	moduleRoot := filepath.Join(repoRoot, "tools", "PSc8y")

	files, err := Generate(specDir)
	if err != nil {
		t.Fatalf("generate: %s", err)
	}

	total := 0
	for _, file := range files {
		outPath := filepath.Join(moduleRoot, file.RelPath)
		if file.Remove {
			if _, err := os.Stat(outPath); err == nil {
				t.Errorf("file of deprecated command should not exist: %s", file.RelPath)
			}
			continue
		}
		total++
		existing, err := os.ReadFile(outPath)
		if err != nil {
			t.Errorf("generated file missing on disk: %s", err)
			continue
		}
		if !bytes.Equal(existing, file.Content) {
			t.Errorf("generated output differs from %s", outPath)
		}
	}

	if total == 0 {
		t.Fatal("no files were generated")
	}
	t.Logf("verified %d generated files", total)
}
