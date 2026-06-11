// Command gen-cli generates the cobra CLI commands (pkg/cmd/**/*.auto.go)
// from the api/spec/json specifications. It replaces the PowerShell build
// pipeline under scripts/build-cli and produces identical output.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/codegen"
)

func main() {
	specDir := flag.String("specs", "api/spec/json", "Directory containing the json specifications")
	outputDir := flag.String("output", "pkg/cmd", "Output directory for the generated commands")
	check := flag.Bool("check", false, "Verify the generated files match the files on disk without writing")
	flag.Parse()

	if err := run(*specDir, *outputDir, *check); err != nil {
		fmt.Fprintf(os.Stderr, "gen-cli: %s\n", err)
		os.Exit(1)
	}
}

func run(specDir, outputDir string, check bool) error {
	specFiles, err := codegen.ListSpecFiles(specDir)
	if err != nil {
		return err
	}
	if len(specFiles) == 0 {
		return fmt.Errorf("no specification files found in %s", specDir)
	}

	changed := 0
	total := 0
	for _, specFile := range specFiles {
		files, warnings, err := codegen.GenerateSpecFile(specFile, outputDir)
		for _, warning := range warnings {
			fmt.Fprintf(os.Stderr, "WARNING: %s file=%s\n", warning.Message, warning.Spec)
		}
		if err != nil {
			return err
		}

		for _, file := range files {
			total++
			outPath := filepath.Join(outputDir, file.RelPath)
			existing, readErr := os.ReadFile(outPath)
			upToDate := readErr == nil && bytes.Equal(existing, file.Source)
			if upToDate {
				continue
			}
			changed++
			if check {
				fmt.Fprintf(os.Stderr, "out of date: %s\n", outPath)
				continue
			}
			if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(outPath, file.Source, 0o644); err != nil {
				return err
			}
		}
	}

	if check {
		if changed > 0 {
			return fmt.Errorf("%d of %d generated files are out of date (run go run ./cmd/gen-cli to update)", changed, total)
		}
		fmt.Printf("%d generated files are up to date\n", total)
		return nil
	}

	fmt.Printf("Generated %d files (%d updated)\n", total, changed)
	return nil
}
