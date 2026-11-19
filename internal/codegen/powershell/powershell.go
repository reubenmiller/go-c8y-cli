// Package powershell generates the PSc8y PowerShell module cmdlets and
// Pester tests from the api/spec/json specifications and packages the
// module. It is a port of the PowerShell scripts under
// scripts/build-powershell and reproduces their output byte-for-byte.
package powershell

import (
	"fmt"
	"os"
	"path"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/codegen"
)

// bom is the UTF-8 byte order mark every generated file starts with (to help
// with encoding in powershell).
const bom = "\xef\xbb\xbf"

// GeneratedFile is one output file operation of a specification. The
// operations apply in order: commands sharing a cmdlet alias overwrite each
// other (the last command of a specification wins), exactly like the
// sequential file writes of the original script.
type GeneratedFile struct {
	// RelPath is the file path relative to the module root (tools/PSc8y),
	// e.g. "Public/Get-Alarm.ps1".
	RelPath string
	// Content is the full file content including BOM and trailing newline.
	// It is nil when Remove is set.
	Content []byte
	// Remove marks the file of a deprecated command for deletion
	// (PowerShell commands do not get deprecation notices).
	Remove bool
}

// GenerateSpec generates the cmdlet and test file operations of one
// specification document.
func GenerateSpec(data []byte) ([]GeneratedFile, error) {
	spec, err := codegen.ParseJSON(data)
	if err != nil {
		return nil, err
	}

	group := spec.Get("group")
	if group.Get("skip").StrEquals("true") {
		return nil, nil
	}
	noun := group.Get("name").Str()

	var files []GeneratedFile
	for _, command := range spec.Get("commands").Items() {
		if command.Get("skip").StrEquals("true") {
			continue
		}
		if command.Get("hidden").Truthy() {
			continue
		}

		cmdletName := command.Get("alias").Get("powershell").Str()
		if cmdletName == "" {
			return nil, fmt.Errorf("command %s is missing the alias.powershell property", command.Get("name").Str())
		}
		deprecated := command.Get("powershell").Get("deprecated").Truthy()

		cmdletPath := path.Join("Public", cmdletName+".ps1")
		testPath := path.Join("Tests", cmdletName+".auto.Tests.ps1")

		if testContent, ok := renderTest(command, cmdletName); ok {
			if deprecated {
				files = append(files, GeneratedFile{RelPath: testPath, Remove: true})
			} else {
				files = append(files, GeneratedFile{RelPath: testPath, Content: []byte(bom + testContent + "\n")})
			}
		}

		if deprecated {
			files = append(files, GeneratedFile{RelPath: cmdletPath, Remove: true})
			continue
		}
		content, err := renderCmdlet(command, noun)
		if err != nil {
			return nil, err
		}
		files = append(files, GeneratedFile{RelPath: cmdletPath, Content: []byte(bom + content + "\n")})
	}

	return files, nil
}

// GenerateSpecFile generates the cmdlet and test file operations of one
// specification file.
func GenerateSpecFile(specPath string) ([]GeneratedFile, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, err
	}
	files, err := GenerateSpec(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", specPath, err)
	}
	return files, nil
}

// Generate generates the file operations of all specification files in a
// directory, reduced to the final state per file (the last operation of a
// path wins).
func Generate(specDir string) ([]GeneratedFile, error) {
	specFiles, err := codegen.ListSpecFiles(specDir)
	if err != nil {
		return nil, err
	}
	if len(specFiles) == 0 {
		return nil, fmt.Errorf("no specification files found in %s", specDir)
	}

	var all []GeneratedFile
	for _, specFile := range specFiles {
		files, err := GenerateSpecFile(specFile)
		if err != nil {
			return nil, err
		}
		all = append(all, files...)
	}

	return reduceFiles(all), nil
}

// reduceFiles collapses a sequence of file operations to the final state per
// path: when several operations target the same RelPath, the last one wins
// (matching the sequential file writes of the original script).
func reduceFiles(all []GeneratedFile) []GeneratedFile {
	last := map[string]int{}
	for i, file := range all {
		last[file.RelPath] = i
	}
	reduced := make([]GeneratedFile, 0, len(last))
	for i, file := range all {
		if last[file.RelPath] == i {
			reduced = append(reduced, file)
		}
	}
	return reduced
}
