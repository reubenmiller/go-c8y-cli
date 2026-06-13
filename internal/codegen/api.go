package codegen

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/imports"
)

// GeneratedFile is one formatted output file of a specification.
type GeneratedFile struct {
	// RelPath is the file path relative to the output directory (pkg/cmd),
	// e.g. "alarms/list/list.auto.go".
	RelPath string
	// Source is the formatted Go source.
	Source []byte
}

// Warning describes a non-fatal generation finding (mirrors the PowerShell
// Write-Warning messages that matter).
type Warning struct {
	Spec    string
	Message string
}

// GenerateSpec generates all output files for one specification document.
// The returned list is empty when the spec is skipped. Formatting matches the
// PowerShell pipeline: root commands are gofmt-ed, subcommands run through
// goimports (which also removes the unused imports of the template).
func GenerateSpec(data []byte, outputDir string) ([]GeneratedFile, []Warning, error) {
	spec, err := ParseJSON(data)
	if err != nil {
		return nil, nil, err
	}

	var warnings []Warning
	group := spec.Get("group")
	groupName := group.Get("name").Str()
	if strings.TrimSpace(groupName) == "" {
		warnings = append(warnings, Warning{Message: "Skipping spec: Specification is missing the information.name property. This is required."})
		return nil, warnings, nil
	}
	if strings.EqualFold(group.Get("skip").Str(), "true") {
		warnings = append(warnings, Warning{Message: "Skipping spec: Specification is marked to be skipped using information.skip property."})
		return nil, warnings, nil
	}

	packageName := goPackageName(groupName)

	var files []GeneratedFile

	rootFile, rootSrc := GenerateRootCommand(spec)
	rootPath := filepath.Join(packageName, rootFile)
	rootFormatted, err := format.Source(rootSrc)
	if err != nil {
		return nil, warnings, fmt.Errorf("gofmt %s: %w", rootPath, err)
	}
	files = append(files, GeneratedFile{RelPath: rootPath, Source: rootFormatted})

	for _, command := range spec.Get("commands").Items() {
		if strings.EqualFold(command.Get("skip").Str(), "true") {
			continue
		}
		subPackageName := goPackageName(command.Get("alias").Get("go").Str())
		cmdFile, cmdSrc := GenerateCommand(command, packageName)
		cmdPath := filepath.Join(packageName, subPackageName, cmdFile)

		// goimports resolves the missing/unused imports relative to the
		// target location so it can discover module-local packages such as
		// pkg/c8ydata.
		formatted, err := imports.Process(filepath.Join(outputDir, cmdPath), cmdSrc, &imports.Options{
			Comments:  true,
			TabIndent: true,
			TabWidth:  8,
		})
		if err != nil {
			return nil, warnings, fmt.Errorf("goimports %s: %w", cmdPath, err)
		}
		files = append(files, GeneratedFile{RelPath: cmdPath, Source: formatted})
	}

	return files, warnings, nil
}

// GenerateSpecFile generates all output files for one specification file.
func GenerateSpecFile(specPath string, outputDir string) ([]GeneratedFile, []Warning, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, nil, err
	}
	files, warnings, err := GenerateSpec(data, outputDir)
	for i := range warnings {
		warnings[i].Spec = specPath
	}
	if err != nil {
		return nil, warnings, fmt.Errorf("%s: %w", specPath, err)
	}
	return files, warnings, nil
}

// ListSpecFiles returns the json specification files of a directory in name
// order.
func ListSpecFiles(specDir string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(specDir, "*.json"))
	if err != nil {
		return nil, err
	}
	return matches, nil
}
