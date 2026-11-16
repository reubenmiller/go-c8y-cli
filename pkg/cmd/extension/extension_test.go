package extension

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestFile(t *testing.T, path string, contents string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte(contents), 0755))
}

func TestResolveExecutable(t *testing.T) {
	dir := filepath.FromSlash("/data/c8y-foo")
	cases := []struct {
		name    string
		binPath string
		goos    string
		want    string
	}{
		{"default", "", "linux", filepath.FromSlash("/data/c8y-foo/bin/c8y-foo")},
		{"default windows", "", "windows", filepath.FromSlash("/data/c8y-foo/bin/c8y-foo.exe")},
		{"custom", "dist/tool", "linux", filepath.FromSlash("/data/c8y-foo/dist/tool")},
		{"custom windows", "dist/tool", "windows", filepath.FromSlash("/data/c8y-foo/dist/tool.exe")},
		{"windows with extension", "dist/tool.cmd", "windows", filepath.FromSlash("/data/c8y-foo/dist/tool.cmd")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, resolveExecutable(dir, tc.binPath, tc.goos))
		})
	}
}

func TestExtensionManifest(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "c8y-foo")
	writeTestFile(t, filepath.Join(dir, "extension.yaml"), `
version: v1
binary:
  mount: tools
  completion: cobra
commands:
  - name: services list
    description: List services
    output: json
  - name: ls
    exec: [services, list]
aliases:
  - name: foo-list
    command: foo services list
    description: example
`)
	ext := &Extension{path: dir}

	manifest, err := ext.Manifest()
	require.NoError(t, err)
	require.NotNil(t, manifest.Binary)
	assert.Equal(t, []string{"tools"}, manifest.Binary.MountPath())
	assert.True(t, manifest.Binary.CompletionEnabled())
	require.Len(t, manifest.Commands, 2)

	assert.Equal(t, []string{"services", "list"}, manifest.Commands[0].Path())
	assert.Equal(t, []string{"services", "list"}, manifest.Commands[0].ExecArgs())
	assert.True(t, manifest.Commands[0].IsJSON())

	assert.Equal(t, []string{"ls"}, manifest.Commands[1].Path())
	assert.Equal(t, []string{"services", "list"}, manifest.Commands[1].ExecArgs())
	assert.False(t, manifest.Commands[1].IsJSON())

	assert.Equal(t, resolveExecutable(dir, "", "linux"), filepath.Join(dir, "bin", "c8y-foo"))
	assert.NotEmpty(t, ext.Executable())

	aliases, err := ext.Aliases()
	require.NoError(t, err)
	require.Len(t, aliases, 1)
	assert.Equal(t, "foo-list", aliases[0].GetName())
}

func TestExtensionManifestIsSharedBetweenCopies(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "c8y-foo")
	writeTestFile(t, filepath.Join(dir, "extension.yaml"), "binary: {}\n")

	ext := Extension{path: dir, manifest: &manifestCache{}}
	copied := ext
	_, err := ext.Manifest()
	require.NoError(t, err)

	// Change the file, the copy should use the cached value
	writeTestFile(t, filepath.Join(dir, "extension.yaml"), "version: v1\n")
	m, err := copied.Manifest()
	require.NoError(t, err)
	assert.NotNil(t, m.Binary)
}

func TestExtensionWithoutManifest(t *testing.T) {
	ext := &Extension{path: t.TempDir()}
	manifest, err := ext.Manifest()
	require.NoError(t, err)
	assert.Nil(t, manifest.Binary)
	assert.Empty(t, manifest.Commands)
	assert.Empty(t, ext.Executable())

	aliases, err := ext.Aliases()
	require.NoError(t, err)
	assert.Empty(t, aliases)
}

func TestExtensionInvalidManifest(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "extension.yaml"), "binary: [\n")
	ext := &Extension{path: dir}
	_, err := ext.Manifest()
	assert.Error(t, err)
	assert.Empty(t, ext.Executable())
}

func TestExtensionKind(t *testing.T) {
	assert.Equal(t, extensions.KindLocal, (&Extension{isLocal: true}).Kind())
	assert.Equal(t, extensions.KindBinary, (&Extension{kind: BinaryKind}).Kind())
	assert.Equal(t, extensions.KindGit, (&Extension{kind: GitKind}).Kind())
}
