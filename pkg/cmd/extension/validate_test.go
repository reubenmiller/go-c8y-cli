package extension

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testGlobalFlags returns some of the c8y global flags
func testGlobalFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("c8y", pflag.ContinueOnError)
	flags.BoolP("verbose", "v", false, "")
	flags.StringP("output", "o", "json", "")
	flags.String("session", "", "")
	flags.Bool("examples", false, "")
	flags.BoolP("force", "f", false, "")
	return flags
}

// binaryScript returns a test binary which lists the given flags for "__complete"
func binaryScript(completeOutput string) string {
	return fmt.Sprintf(`#!/bin/sh
if [ "$1" = "__complete" ]; then
  printf '%%s\n' %s
  echo ":4"
  exit 0
fi
echo "binary $*"
`, completeOutput)
}

type testExt struct {
	t   *testing.T
	dir string
}

func newTestExt(t *testing.T, name string) *testExt {
	t.Helper()
	return &testExt{t: t, dir: filepath.Join(t.TempDir(), name)}
}

func (e *testExt) file(rel, contents string, mode os.FileMode) *testExt {
	e.t.Helper()
	path := filepath.Join(e.dir, filepath.FromSlash(rel))
	require.NoError(e.t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(e.t, os.WriteFile(path, []byte(contents), mode))
	return e
}

func (e *testExt) validate(opts ValidateOptions) []Finding {
	e.t.Helper()
	if opts.GlobalFlags == nil {
		opts.GlobalFlags = testGlobalFlags()
	}
	return ValidateExtension(&Extension{path: e.dir, isLocal: true, manifest: &manifestCache{}}, opts)
}

func skipWithoutShell(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("uses shell scripts as test binaries")
	}
}

func findingsText(findings []Finding) string {
	lines := []string{}
	for _, f := range findings {
		lines = append(lines, fmt.Sprintf("%s|%s|%s|%s", f.Severity, f.Check, f.Location, f.Message))
	}
	return strings.Join(lines, "\n")
}

func TestValidateCleanExtension(t *testing.T) {
	skipWithoutShell(t)
	ext := newTestExt(t, "c8y-clean").
		file("extension.yaml", "version: v1\nbinary:\n  completion: cobra\ncommands:\n  - name: services list\n    output: json\naliases:\n  - name: x\n    command: devices list\n    description: x\n", 0644).
		file("bin/c8y-clean", binaryScript("--status --help -h --examples"), 0755).
		file("api/devices.yaml", "group:\n  name: devices\n  description: Devices\ncommands:\n  - name: list\n    description: List\n    method: GET\n    path: inventory/managedObjects\n", 0644).
		file("commands/hello", "#!/bin/sh\necho hello\n", 0755)

	findings := ext.validate(ValidateOptions{})
	assert.Empty(t, findings, findingsText(findings))
}

func TestValidateFlagClashes(t *testing.T) {
	skipWithoutShell(t)
	ext := newTestExt(t, "c8y-clash").
		file("extension.yaml", "binary:\n  completion: cobra\ncommands:\n  - name: services list\n", 0644).
		file("bin/c8y-clash", binaryScript("--verbose -v --output -o --status --help -h"), 0755)

	findings := ext.validate(ValidateOptions{})
	require.Len(t, findings, 2, findingsText(findings))
	assert.Equal(t, SeverityWarning, findings[0].Severity)
	assert.Equal(t, CheckFlags, findings[0].Check)
	assert.Equal(t, "c8y clash services list", findings[0].Location)
	assert.Contains(t, findings[0].Message, "--verbose, -v")
	assert.Contains(t, findings[0].Message, "-v/--verbose")
	assert.Contains(t, findings[1].Message, "--output, -o")

	// The flag check can be skipped (e.g. after install)
	assert.Empty(t, ext.validate(ValidateOptions{SkipFlagCheck: true}))
}

func TestValidateFlagsNotChecked(t *testing.T) {
	skipWithoutShell(t)
	ext := newTestExt(t, "c8y-nocobra").
		file("extension.yaml", "binary: {}\ncommands:\n  - name: list\n", 0644).
		file("bin/c8y-nocobra", "#!/bin/sh\n", 0755)

	findings := ext.validate(ValidateOptions{})
	require.Len(t, findings, 1, findingsText(findings))
	assert.Equal(t, SeverityInfo, findings[0].Severity)
	assert.Contains(t, findings[0].Message, "not checked")
}

func TestValidateFlagsCompleteFails(t *testing.T) {
	skipWithoutShell(t)
	ext := newTestExt(t, "c8y-broken").
		file("extension.yaml", "binary: {completion: cobra}\ncommands:\n  - name: list\n", 0644).
		file("bin/c8y-broken", "#!/bin/sh\nexit 3\n", 0755)
	findings := ext.validate(ValidateOptions{})
	require.Len(t, findings, 1, findingsText(findings))
	assert.Equal(t, SeverityWarning, findings[0].Severity)
	assert.Contains(t, findings[0].Message, "could not read the flags")

	ext = newTestExt(t, "c8y-slow").
		file("extension.yaml", "binary: {completion: cobra}\ncommands:\n  - name: list\n", 0644).
		file("bin/c8y-slow", "#!/bin/sh\nexec sleep 5\n", 0755)
	findings = ext.validate(ValidateOptions{CompleteTimeout: 100 * time.Millisecond})
	require.Len(t, findings, 1, findingsText(findings))
	assert.Contains(t, findings[0].Message, "timed out")
}

func TestValidateManifest(t *testing.T) {
	skipWithoutShell(t)
	ext := newTestExt(t, "c8y-manifest").
		file("extension.yaml", `version: v1
comands:
  - name: typo
binary:
  completion: bash
  pth: bin/x
commands:
  - name: list
    output: xml
    exec: [list]
  - name: list
  - name: "  "
`, 0644).
		file("bin/c8y-manifest", "#!/bin/sh\n", 0755)

	text := findingsText(ext.validate(ValidateOptions{SkipFlagCheck: true}))
	assert.Contains(t, text, `warning|manifest|extension.yaml:2|unknown key 'comands'`)
	assert.Contains(t, text, `warning|manifest|extension.yaml:6|unknown key 'binary.pth'`)
	assert.Contains(t, text, `error|manifest|extension.yaml|invalid binary.completion 'bash'`)
	assert.Contains(t, text, `error|manifest|extension.yaml commands[0]|invalid output 'xml'`)
	assert.Contains(t, text, `error|manifest|extension.yaml commands[1]|duplicate command 'list'`)
	assert.Contains(t, text, `error|manifest|extension.yaml commands[2]|the command name is empty`)

	// errors are listed before warnings
	findings := ext.validate(ValidateOptions{SkipFlagCheck: true})
	assert.Equal(t, SeverityError, findings[0].Severity)
}

func TestValidateInvalidManifest(t *testing.T) {
	ext := newTestExt(t, "c8y-invalid").file("extension.yaml", "binary: [\n", 0644)
	findings := ext.validate(ValidateOptions{})
	require.Len(t, findings, 1, findingsText(findings))
	assert.Equal(t, SeverityError, findings[0].Severity)
	assert.Contains(t, findings[0].Message, "invalid yaml")
}

func TestValidateCommandsWithoutBinary(t *testing.T) {
	ext := newTestExt(t, "c8y-nobinary").file("extension.yaml", "commands:\n  - name: list\n  - name: get\n", 0644)
	findings := ext.validate(ValidateOptions{})
	require.Len(t, findings, 1, findingsText(findings))
	assert.Equal(t, SeverityError, findings[0].Severity)
	assert.Contains(t, findings[0].Message, "2 declared command(s) are ignored")
}

func TestValidateBinary(t *testing.T) {
	ext := newTestExt(t, "c8y-missing").file("extension.yaml", "binary: {}\n", 0644)
	findings := ext.validate(ValidateOptions{})
	require.Len(t, findings, 1, findingsText(findings))
	assert.Equal(t, "error|binary|bin/c8y-missing|the binary does not exist, so the binary commands are not available. Build it, or check binary.path", findingsText(findings))

	ext = newTestExt(t, "c8y-noexec").
		file("extension.yaml", "binary: {}\n", 0644).
		file("bin/c8y-noexec", "#!/bin/sh\n", 0644)
	findings = ext.validate(ValidateOptions{GOOS: "linux"})
	require.Len(t, findings, 1, findingsText(findings))
	assert.Contains(t, findings[0].Message, "not executable")

	// Not checked on Windows
	assert.Empty(t, ext.validate(ValidateOptions{GOOS: "windows"}))
}

func TestValidateCommandConflicts(t *testing.T) {
	skipWithoutShell(t)
	ext := newTestExt(t, "c8y-conflict").
		file("extension.yaml", "binary: {}\ncommands:\n  - name: management logs\n", 0644).
		file("bin/c8y-conflict", "#!/bin/sh\n", 0755).
		file("api/management.yaml", "group:\n  name: management\n  description: Manage\ncommands:\n  - name: health\n    description: Health\n    method: GET\n    path: health\n", 0644).
		file("commands/management/health", "#!/bin/sh\n", 0755).
		file("commands/management/logs", "#!/bin/sh\n", 0755)

	text := findingsText(ext.validate(ValidateOptions{SkipFlagCheck: true}))
	assert.Contains(t, text, `warning|commands|c8y conflict management health|commands/management/health is ignored, as api/management.yaml defines the same command`)
	assert.Contains(t, text, `warning|commands|c8y conflict management logs|commands/management/logs is ignored, as extension.yaml defines the same command`)
}

func TestValidateAPIAndScripts(t *testing.T) {
	ext := newTestExt(t, "c8y-files").
		file("api/broken.yaml", "group: [\n", 0644).
		file("api/README.md", "not a spec", 0644).
		file("commands/notexec", "#!/bin/sh\n", 0644)

	text := findingsText(ext.validate(ValidateOptions{GOOS: "linux"}))
	assert.Contains(t, text, "error|api|api/broken.yaml|invalid api specification")
	assert.Contains(t, text, "warning|scripts|commands/notexec|the script is not executable")
	assert.NotContains(t, text, "README")

	assert.NotContains(t, findingsText(ext.validate(ValidateOptions{GOOS: "windows"})), "not executable")
}

func TestValidateKitchensinkFixture(t *testing.T) {
	dir, err := filepath.Abs(filepath.Join("..", "..", "..", "tests", "testdata", "extensions", "c8y-kitchensink"))
	require.NoError(t, err)
	findings := ValidateExtension(&Extension{path: dir, isLocal: true, manifest: &manifestCache{}}, ValidateOptions{GlobalFlags: testGlobalFlags()})
	errorCount, _ := CountFindings(findings)
	assert.Zero(t, errorCount, findingsText(findings))
}

func TestCountFindings(t *testing.T) {
	errorCount, warningCount := CountFindings([]Finding{
		{Severity: SeverityError}, {Severity: SeverityWarning}, {Severity: SeverityWarning}, {Severity: SeverityInfo},
	})
	assert.Equal(t, 1, errorCount)
	assert.Equal(t, 2, warningCount)
}
