package codegen

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestGoldenSpecOutput verifies that the generator runs for every
// api/spec/json specification and reproduces every committed
// pkg/cmd/**/*.auto.go file byte-for-byte. Commands which are migrated to
// hand-written commands (no *.auto.go on disk, see IsMigrated) are skipped.
func TestGoldenSpecOutput(t *testing.T) {
	repoRoot := "../.."
	specDir := filepath.Join(repoRoot, "api", "spec", "json")
	outputDir := filepath.Join(repoRoot, "pkg", "cmd")

	specFiles, err := ListSpecFiles(specDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(specFiles) == 0 {
		t.Fatalf("no specification files found in %s", specDir)
	}

	generated := map[string]bool{}
	total := 0
	migrated := 0
	for _, specFile := range specFiles {
		files, _, err := GenerateSpecFile(specFile, outputDir)
		if err != nil {
			t.Fatalf("generate %s: %s", specFile, err)
		}
		for _, file := range files {
			total++
			generated[filepath.ToSlash(file.RelPath)] = true
			if IsMigrated(outputDir, file) {
				// hand-written command (go-c8y v2 Runner), see IsMigrated
				migrated++
				continue
			}
			outPath := filepath.Join(outputDir, file.RelPath)
			existing, err := os.ReadFile(outPath)
			if err != nil {
				t.Errorf("%s: read generated file: %s", specFile, err)
				continue
			}
			if !bytes.Equal(existing, file.Source) {
				t.Errorf("%s: generated output differs from %s", specFile, outPath)
			}
		}
	}

	if total == 0 {
		t.Fatal("no files were generated")
	}
	t.Logf("generated %d files from the specs, %d of them are migrated to hand-written commands", total, migrated)

	// Stale *.auto.go files no longer produced from the specs are reported so
	// they are not silently shipped
	err = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !bytes.HasSuffix([]byte(path), []byte(".auto.go")) {
			return nil
		}
		rel, err := filepath.Rel(outputDir, path)
		if err != nil {
			return err
		}
		if !generated[filepath.ToSlash(rel)] {
			t.Logf("stale file not generated from any spec: %s", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
