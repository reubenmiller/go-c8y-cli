package extension

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newScaffoldManager returns a manager which can create extensions. Only git is available,
// so the Go dependencies are not downloaded during create (the SDK version may not be released yet)
func newScaffoldManager(t *testing.T) *Manager {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	ios, _, _, _ := iostreams.Test()
	return &Manager{
		lookPath: func(name string) (string, error) {
			if name == "git" {
				return exec.LookPath("git")
			}
			return "", exec.ErrNotFound
		},
		newCommand: exec.Command,
		io:         ios,
	}
}

func runIn(t *testing.T, dir string, env []string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s %v: %s", name, args, out)
	return string(out)
}

// localSDKDir returns the path to the extension SDK in this repository
func localSDKDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "extsdk"))
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(dir, "go.mod"))
	return dir
}

// scaffoldGoExtension creates the go template, points it at the local SDK, and builds it
func scaffoldGoExtension(t *testing.T, name string) string {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping scaffold build in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
	sdkDir := localSDKDir(t) // before changing the working directory
	m := newScaffoldManager(t)
	t.Chdir(t.TempDir())
	require.NoError(t, m.Create(name, extensions.GoBinTemplateType))

	dir, err := filepath.Abs(name)
	require.NoError(t, err)
	runIn(t, dir, nil, "go", "mod", "edit", "-replace=github.com/reubenmiller/go-c8y-cli/pkg/extsdk="+sdkDir)
	runIn(t, dir, nil, "go", "mod", "tidy")
	runIn(t, dir, nil, "go", "build", "-o", filepath.Join("bin", name), ".")
	return dir
}

func TestCreateGoExtension(t *testing.T) {
	dir := scaffoldGoExtension(t, "c8y-gofoo")

	for _, file := range []string{"go.mod", "main.go", "extension.yaml", "views/services.json", ".goreleaser.yaml", ".github/workflows/release.yml", "README.md", ".gitignore"} {
		assert.FileExists(t, filepath.Join(dir, filepath.FromSlash(file)))
	}

	// The extension is usable as a local extension
	ext := &Extension{path: dir}
	manifest, err := ext.Manifest()
	require.NoError(t, err)
	require.NotNil(t, manifest.Binary)
	assert.True(t, manifest.Binary.CompletionEnabled())
	require.Len(t, manifest.Commands, 1)
	assert.Equal(t, "services list", manifest.Commands[0].Name)
	assert.True(t, manifest.Commands[0].IsJSON())
	require.FileExists(t, ext.Executable())

	// The example command writes json when run by c8y
	out := runIn(t, dir, []string{"C8Y_EXTENSION_CONTEXT_VERSION=1", "C8Y_EXTENSION_NAME=gofoo"}, ext.Executable(), "services", "list")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], `"name":"ingest"`)
	assert.Contains(t, lines[0], `"type":"c8ygofoo_service"`)

	// and fails with a clear error when not run by c8y
	cmd := exec.Command(ext.Executable(), "services", "list")
	cmd.Env = []string{}
	output, err := cmd.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(output), "must be run through c8y")

	// The generated extension declares its name and has no problems
	declared, err := readDeclaredName(dir)
	require.NoError(t, err)
	assert.Equal(t, "gofoo", declared)
	findings := ValidateExtension(&Extension{path: dir, isLocal: true, manifest: &manifestCache{}}, ValidateOptions{CompleteTimeout: testCompleteTimeout, GlobalFlags: testGlobalFlags(), RootCommand: testRootCommand()})
	assert.Empty(t, findings, findingsText(findings))

	// The release workflow validates the extension
	workflow, err := os.ReadFile(filepath.Join(dir, ".github", "workflows", "release.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(workflow), "extensions validate . --strict")
	assert.Contains(t, string(workflow), "go build -o bin/c8y-gofoo .")

	// The view matches the example items
	view, err := os.ReadFile(filepath.Join(dir, "views", "services.json"))
	require.NoError(t, err)
	assert.Contains(t, string(view), `"type": "c8ygofoo_service"`)
}

func TestScaffoldReleaseArchivesMatchInstallFormat(t *testing.T) {
	if _, err := exec.LookPath("goreleaser"); err != nil {
		t.Skip("goreleaser is not installed")
	}
	name := "c8y-relfoo"
	dir := scaffoldGoExtension(t, name)

	gitEnv := []string{"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.invalid", "GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.invalid"}
	runIn(t, dir, gitEnv, "git", "add", "-A")
	runIn(t, dir, gitEnv, "git", "commit", "--quiet", "-m", "init")
	runIn(t, dir, gitEnv, "git", "tag", "v1.0.0")
	runIn(t, dir, nil, "goreleaser", "release", "--snapshot", "--clean", "--skip=publish")

	entries, err := os.ReadDir(filepath.Join(dir, "dist"))
	require.NoError(t, err)
	assets := []releaseAsset{}
	for _, entry := range entries {
		if !entry.IsDir() {
			assets = append(assets, releaseAsset{Name: entry.Name()})
		}
	}
	require.NotNil(t, findChecksumAsset(assets))

	checksumFile, err := os.Open(filepath.Join(dir, "dist", "checksums.txt"))
	require.NoError(t, err)
	defer checksumFile.Close()
	checksums, err := parseChecksums(checksumFile)
	require.NoError(t, err)

	for _, platform := range []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64", "windows/arm64"} {
		goos, goarch, _ := strings.Cut(platform, "/")
		t.Run(platform, func(t *testing.T) {
			asset, kind, _ := selectAsset(assets, goos, goarch, false)
			require.NotNil(t, asset, "no asset selected for %s", platform)
			assert.True(t, kind.IsArchive())

			path := filepath.Join(dir, "dist", asset.Name)
			sum, err := fileSHA256(path)
			require.NoError(t, err)
			assert.Equal(t, checksums[asset.Name], sum)

			content := filepath.Join(t.TempDir(), "content")
			require.NoError(t, extractArchive(path, kind, content))
			root, err := extensionRoot(content)
			require.NoError(t, err)

			manifest, err := readManifest(root)
			require.NoError(t, err)
			require.NotNil(t, manifest.Binary)
			assert.FileExists(t, resolveExecutable(root, manifest.Binary.Path, goos))
			assert.FileExists(t, filepath.Join(root, "views", "services.json"))
		})
	}
}

func TestCreateOtherExtension(t *testing.T) {
	m := newScaffoldManager(t)
	t.Chdir(t.TempDir())
	require.NoError(t, m.Create("c8y-otherfoo", extensions.OtherBinTemplateType))

	for _, file := range []string{"extension.yaml", "script/build.sh", ".github/workflows/release.yml", "README.md", ".gitignore"} {
		assert.FileExists(t, filepath.Join("c8y-otherfoo", filepath.FromSlash(file)))
	}
	assert.NoFileExists(t, filepath.Join("c8y-otherfoo", "main.go"))

	manifest, err := readManifest("c8y-otherfoo")
	require.NoError(t, err)
	require.NotNil(t, manifest.Binary)
	assert.Equal(t, "bin/c8y-otherfoo", manifest.Binary.Path)
	assert.Equal(t, "otherfoo", manifest.Name)

	out := runIn(t, "c8y-otherfoo", nil, "git", "ls-files", "-s", "script/build.sh")
	assert.True(t, strings.HasPrefix(out, "100755"), "build.sh should be executable in git: %s", out)

	readme, err := os.ReadFile(filepath.Join("c8y-otherfoo", "README.md"))
	require.NoError(t, err)
	assert.Contains(t, string(readme), "c8y-otherfoo_1.0.0_linux_amd64.tar.gz")
}

func TestCreateScriptExtensionIsDefault(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses git chmod")
	}
	m := newScaffoldManager(t)
	t.Chdir(t.TempDir())
	require.NoError(t, m.Create("c8y-scriptfoo", extensions.GitTemplateType))

	assert.FileExists(t, filepath.Join("c8y-scriptfoo", "commands", "services", "list"))
	manifest, err := readManifest("c8y-scriptfoo")
	require.NoError(t, err)
	assert.Nil(t, manifest.Binary)
	assert.Equal(t, "scriptfoo", manifest.Name)

	findings := ValidateExtension(&Extension{path: "c8y-scriptfoo", isLocal: true, manifest: &manifestCache{}}, ValidateOptions{SkipFlagCheck: true, RootCommand: testRootCommand()})
	for _, f := range findings {
		assert.NotEqual(t, CheckName, f.Check, findingsText(findings))
		assert.NotEqual(t, CheckManifest, f.Check, findingsText(findings))
	}
}
