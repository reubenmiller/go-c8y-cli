// Command gen-powershell generates the PSc8y PowerShell module cmdlets
// (tools/PSc8y/Public) and Pester tests (tools/PSc8y/Tests) and packages the
// module under tools/PSc8y/dist.
//
// By default it reads the api/spec/json specifications (the legacy, REST-spec
// pipeline). With --from-tree it instead projects the cmdlets from the live
// c8y command tree — the hand-written commands are the source of truth — which
// is the target architecture (see proposals/CLI_CODEGEN_INVERSION.md). During
// migration both inputs coexist; --command scopes the tree projection to a
// subtree so it can be generated and diffed without touching the spec-generated
// cmdlets.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/clibuild"
	"github.com/reubenmiller/go-c8y-cli/v2/internal/codegen/powershell"
)

func main() {
	specDir := flag.String("specs", "api/spec/json", "Directory containing the json specifications")
	moduleDir := flag.String("module", "tools/PSc8y", "PSc8y module root directory (output)")
	check := flag.Bool("check", false, "Verify the generated files match the files on disk without writing (skips packaging)")
	skipPackage := flag.Bool("skip-package", false, "Only generate the cmdlets and tests, do not package the module under dist")
	fromTree := flag.Bool("from-tree", false, "Project cmdlets/tests from the live command tree instead of the REST spec")
	command := flag.String("command", "", "With --from-tree, limit generation to this command subtree (e.g. \"devices\")")
	flag.Parse()

	if err := run(*specDir, *moduleDir, *command, *check, *skipPackage, *fromTree); err != nil {
		fmt.Fprintf(os.Stderr, "gen-powershell: %s\n", err)
		os.Exit(1)
	}
}

func run(specDir, moduleDir, command string, check, skipPackage, fromTree bool) error {
	var files []powershell.GeneratedFile
	var err error

	if fromTree {
		var root, buildErr = clibuild.NewRootCommand()
		if buildErr != nil {
			return buildErr
		}
		root.InitDefaultHelpCmd()
		files, err = powershell.GenerateFromTree(root.Command, command)
		if err != nil {
			return err
		}
		// A tree projection is partial during migration; packaging the module
		// from a subtree is meaningless, so never package in this mode.
		skipPackage = true
	} else {
		files, err = powershell.Generate(specDir)
		if err != nil {
			return err
		}
	}

	changed := 0
	total := 0
	for _, file := range files {
		outPath := filepath.Join(moduleDir, file.RelPath)

		// Remove files of deprecated commands as powershell commands do not
		// get deprecation notices
		if file.Remove {
			if _, err := os.Stat(outPath); err != nil {
				continue
			}
			changed++
			if check {
				fmt.Fprintf(os.Stderr, "out of date (deprecated, must be removed): %s\n", outPath)
				continue
			}
			if err := os.Remove(outPath); err != nil {
				return err
			}
			continue
		}

		total++
		existing, readErr := os.ReadFile(outPath)
		upToDate := readErr == nil && bytes.Equal(existing, file.Content)
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
		if err := os.WriteFile(outPath, file.Content, 0o644); err != nil {
			return err
		}
	}

	if check {
		if changed > 0 {
			return fmt.Errorf("%d of %d generated files are out of date (run go run ./cmd/gen-powershell to update)", changed, total)
		}
		fmt.Printf("%d generated files are up to date\n", total)
		return nil
	}

	fmt.Printf("Generated %d files (%d updated)\n", total, changed)

	if skipPackage {
		return nil
	}

	version, err := powershell.ResolveVersion(moduleDir)
	if err != nil {
		return err
	}
	exportPath, err := powershell.BuildModule(moduleDir, version)
	if err != nil {
		return err
	}
	fmt.Printf("\n    Created module in: %s\n\n", exportPath)
	return nil
}
