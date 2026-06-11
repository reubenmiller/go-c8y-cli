package powershell

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// This file ports tools/PSc8y/tools/build.psm1 (Export-ProductionModule):
// it builds the distributable module under tools/PSc8y/dist consisting of
// the concatenated PSc8y.psm1, the version stamped PSc8y.psd1, the module
// resources and the PSc8y.zip archive.
//
// One intentional deviation: the PowerShell pipeline stamps the version with
// Update-ModuleManifest, which rewrites the whole manifest in its own
// format. The Go port instead substitutes the ModuleVersion value (and
// inserts the Prerelease entry) in place, keeping the source manifest
// formatting. The resulting manifest is functionally equivalent.

// moduleName is the name of the generated PowerShell module.
const moduleName = "PSc8y"

// manualExclusions are the Public-manual functions excluded from the
// published module.
var manualExclusions = map[string]bool{
	"New-TestMicroservice.ps1":      true,
	"New-TestHostedApplication.ps1": true,
}

// GitVersion is the module version derived from git describe (the port of
// Get-GitVersion).
type GitVersion struct {
	Version      string
	Prerelease   string
	IsPrerelease bool
}

var versionPattern = regexp.MustCompile(`^v?\d+`)

// ResolveVersion ports Get-GitVersion: git describe with a GITHUB_REF
// fallback, defaulting to 0.0.0.
func ResolveVersion(dir string) (GitVersion, error) {
	version := ""
	cmd := exec.Command("git", "describe")
	cmd.Dir = dir
	if out, err := cmd.Output(); err == nil {
		version = strings.TrimSpace(string(out))
	} else if ref := os.Getenv("GITHUB_REF"); ref != "" {
		version = ref[strings.LastIndex(ref, "/")+1:]
	}

	if !versionPattern.MatchString(version) {
		version = "0.0.0"
	}

	parts := strings.SplitN(version, "-", 2)
	v := GitVersion{
		Version: strings.TrimPrefix(parts[0], "v"),
	}
	if len(parts) > 1 && parts[1] != "" {
		v.IsPrerelease = true
		v.Prerelease = strings.SplitN(parts[1], "-", 2)[0]
	}
	return v, nil
}

// BuildModule ports Export-ProductionModule: it updates the exported
// function list of the source manifest, assembles the distributable module
// under <moduleRoot>/dist/PSc8y and packages it as <moduleRoot>/dist/PSc8y.zip.
// It returns the path of the created module directory.
func BuildModule(moduleRoot string, version GitVersion) (string, error) {
	publicFiles, err := publicFunctionFiles(moduleRoot)
	if err != nil {
		return "", err
	}
	functionNames := make([]string, 0, len(publicFiles))
	for _, file := range publicFiles {
		functionNames = append(functionNames, strings.TrimSuffix(filepath.Base(file), ".ps1"))
	}

	manifestPath := filepath.Join(moduleRoot, moduleName+".psd1")
	if err := updateManifestFunctions(manifestPath, functionNames); err != nil {
		return "", err
	}

	// Publish-ModuleArtifacts
	artifactRoot := filepath.Join(moduleRoot, "dist")
	if err := os.RemoveAll(artifactRoot); err != nil {
		return "", err
	}
	moduleDist := filepath.Join(artifactRoot, moduleName)
	if err := os.MkdirAll(moduleDist, 0o755); err != nil {
		return "", err
	}

	// Copy the module into the dist folder
	if err := copyDir(filepath.Join(moduleRoot, "Dependencies"), filepath.Join(moduleDist, "Dependencies"), "c8y"); err != nil {
		return "", err
	}
	if err := copyDir(filepath.Join(moduleRoot, "format-data"), filepath.Join(moduleDist, "format-data"), ""); err != nil {
		return "", err
	}
	if err := copyDir(filepath.Join(moduleRoot, "Templates"), filepath.Join(moduleDist, "Templates"), ""); err != nil {
		return "", err
	}
	if err := copyFile(manifestPath, filepath.Join(moduleDist, moduleName+".psd1")); err != nil {
		return "", err
	}

	// Construct the distributed .psm1 file
	if err := writeModulePSM1(moduleRoot, publicFiles, functionNames, filepath.Join(moduleDist, moduleName+".psm1")); err != nil {
		return "", err
	}

	// Update version
	if err := stampManifestVersion(filepath.Join(moduleDist, moduleName+".psd1"), version); err != nil {
		return "", err
	}

	// Package the module in /dist
	if err := zipDir(artifactRoot, filepath.Join(artifactRoot, moduleName+".zip")); err != nil {
		return "", err
	}

	return moduleDist, nil
}

// publicFunctionFiles returns the function files included in the module:
// Public/*.ps1 followed by Public-manual/*.ps1 without the excluded entries.
func publicFunctionFiles(moduleRoot string) ([]string, error) {
	public, err := listPS1Files(filepath.Join(moduleRoot, "Public"))
	if err != nil {
		return nil, err
	}
	if len(public) == 0 {
		return nil, fmt.Errorf("no public functions found in %s", filepath.Join(moduleRoot, "Public"))
	}
	manual, err := listPS1Files(filepath.Join(moduleRoot, "Public-manual"))
	if err != nil {
		return nil, err
	}
	files := public
	for _, file := range manual {
		if !manualExclusions[filepath.Base(file)] {
			files = append(files, file)
		}
	}
	return files, nil
}

// listPS1Files returns the *.ps1 files of a directory sorted by name the way
// Get-ChildItem does (case-insensitive). A missing directory yields an empty
// list (Get-ChildItem -ErrorAction SilentlyContinue).
func listPS1Files(dir string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.ps1"))
	if err != nil {
		return nil, err
	}
	sort.Slice(matches, func(a, b int) bool {
		left := strings.ToLower(matches[a])
		right := strings.ToLower(matches[b])
		if left == right {
			return matches[a] < matches[b]
		}
		return left < right
	})
	return matches, nil
}

// fileContent reads a file the way Get-Content does: BOM stripped, line
// endings normalized to \n and exactly one trailing newline.
func fileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := strings.TrimPrefix(string(data), "\uFEFF")
	content = strings.ReplaceAll(content, "\r\n", "\n")
	return strings.TrimSuffix(content, "\n") + "\n", nil
}

// writeModulePSM1 ports New-ModulePSMFile: PartOne, the function regions,
// the Export-ModuleMember statement and PartTwo concatenated into the
// distributed .psm1 file.
func writeModulePSM1(moduleRoot string, publicFiles, functionNames []string, outPath string) error {
	b := &bytes.Buffer{}
	b.WriteString(bom)

	partOne, err := fileContent(filepath.Join(moduleRoot, "tools", "modulefile", "PartOne.ps1"))
	if err != nil {
		return err
	}
	b.WriteString(partOne)

	privateFiles, err := listPS1Files(filepath.Join(moduleRoot, "Private"))
	if err != nil {
		return err
	}
	enumFiles, err := listPS1Files(filepath.Join(moduleRoot, "Enums"))
	if err != nil {
		return err
	}
	completionFiles, err := listPS1Files(filepath.Join(moduleRoot, "completions"))
	if err != nil {
		return err
	}
	utilityFiles, err := listPS1Files(filepath.Join(moduleRoot, "utilities"))
	if err != nil {
		return err
	}

	regions := []struct {
		name  string
		files []string
	}{
		{"Private Functions", privateFiles},
		{"Public Functions", publicFiles},
		{"Enums", enumFiles},
		{"Completions", completionFiles},
		{"Utilities", utilityFiles},
	}
	for i, region := range regions {
		if i == 0 {
			b.WriteString("\n")
		}
		b.WriteString("#region " + region.name + "\n")
		for _, file := range region.files {
			content, err := fileContent(file)
			if err != nil {
				return err
			}
			// Get-Content | Out-String | Out-File appends an extra newline
			b.WriteString(content + "\n")
		}
		b.WriteString("#endregion\n\n")
	}

	b.WriteString("Export-ModuleMember -Function " + strings.Join(functionNames, ",") + "\n\n")

	partTwo, err := fileContent(filepath.Join(moduleRoot, "tools", "modulefile", "PartTwo.ps1"))
	if err != nil {
		return err
	}
	b.WriteString(partTwo)

	return os.WriteFile(outPath, b.Bytes(), 0o644)
}

// updateManifestFunctions ports Update-ModuleManifestFunctions: the
// FunctionsToExport block of the manifest (everything from the
// "FunctionsToExport =" line up to the "VariablesToExport" line) is replaced
// with the explicit function name list.
func updateManifestFunctions(manifestPath string, functionNames []string) error {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\r\n"), "\n")

	start := -1
	end := -1
	for i, line := range lines {
		if start < 0 && strings.Contains(line, "FunctionsToExport =") {
			start = i
		}
		if end < 0 && strings.Contains(line, "VariablesToExport =") {
			end = i
		}
	}
	if start < 0 || end < 0 || end <= start {
		return fmt.Errorf("%s: could not locate the FunctionsToExport block", manifestPath)
	}

	block := make([]string, 0, len(functionNames)+1)
	block = append(block, "FunctionsToExport = @(")
	for i, name := range functionNames {
		entry := "\t'" + name + "'"
		if i < len(functionNames)-1 {
			entry += ","
		} else {
			entry += ")"
		}
		block = append(block, entry)
	}

	updated := make([]string, 0, len(lines))
	updated = append(updated, lines[:start]...)
	updated = append(updated, block...)
	updated = append(updated, lines[end:]...)

	return os.WriteFile(manifestPath, []byte(strings.Join(updated, "\n")+"\n"), 0o644)
}

// stampManifestVersion sets the ModuleVersion and (for prerelease builds)
// inserts the Prerelease entry of the PSData section. It also applies the
// export normalizations Update-ModuleManifest used to do: an explicit empty
// CmdletsToExport (the module ships no cmdlets, and PowerShellGet warns
// about implicit wildcard exports when publishing) and no empty
// DefaultCommandPrefix.
func stampManifestVersion(manifestPath string, version GitVersion) error {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	content := string(data)

	reModuleVersion := regexp.MustCompile(`(?m)^ModuleVersion = '[^']*'`)
	if !reModuleVersion.MatchString(content) {
		return fmt.Errorf("%s: could not locate the ModuleVersion entry", manifestPath)
	}
	content = reModuleVersion.ReplaceAllString(content, "ModuleVersion = '"+version.Version+"'")

	if !strings.Contains(content, "CmdletsToExport") {
		variablesLine := "# VariablesToExport = '*'\n"
		content = strings.Replace(content, variablesLine,
			variablesLine+"\n# Cmdlets to export from this module, for best performance, do not use wildcards and do not delete the entry, use an empty array if there are no cmdlets to export.\nCmdletsToExport = @()\n", 1)
	}
	content = strings.Replace(content, "\nDefaultCommandPrefix = ''", "\n# DefaultCommandPrefix = ''", 1)

	if version.IsPrerelease {
		marker := "        # External dependent modules of this module"
		if !strings.Contains(content, marker) {
			return fmt.Errorf("%s: could not locate the PSData section to add the Prerelease entry", manifestPath)
		}
		prerelease := "        # Prerelease string of this module\n" +
			"        Prerelease = '" + version.Prerelease + "'\n\n"
		content = strings.Replace(content, marker, prerelease+marker, 1)
	}

	return os.WriteFile(manifestPath, []byte(content), 0o644)
}

// copyFile copies a single file.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// copyDir copies a directory tree. When prefix is set, only the entries
// whose name starts with the prefix are copied (Copy-Item -Filter "c8y*").
// The destination directory is created even when no entry matches.
func copyDir(src, dst string, prefix string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if prefix != "" && !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath, ""); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

// zipDir packages the directory content (excluding the archive itself) the
// way System.IO.Compression.ZipFile::CreateFromDirectory does: file entries
// with forward slash paths relative to the directory, plus entries for empty
// directories.
func zipDir(dir, zipPath string) error {
	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	w := zip.NewWriter(out)

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == dir || path == zipPath {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)

		if info.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			if len(entries) > 0 {
				return nil
			}
			header := &zip.FileHeader{Name: name + "/", Method: zip.Deflate}
			header.Modified = info.ModTime()
			_, err = w.CreateHeader(header)
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = name
		header.Method = zip.Deflate
		entry, err := w.CreateHeader(header)
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		_, err = io.Copy(entry, in)
		return err
	})
	if err != nil {
		_ = w.Close()
		_ = out.Close()
		return err
	}
	if err := w.Close(); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
