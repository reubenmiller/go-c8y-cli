package extension

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/ghrepo"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// fakeGitHub is a fake GitHub release API
type fakeGitHub struct {
	mu       sync.Mutex
	releases map[string]*fakeRelease // by tag
	latest   string
	requests []string
}

type fakeRelease struct {
	tag    string
	assets map[string][]byte
}

func newFakeGitHub() *fakeGitHub {
	return &fakeGitHub{releases: map[string]*fakeRelease{}}
}

func (g *fakeGitHub) addRelease(tag string, assets map[string][]byte, checksums bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if checksums {
		lines := []string{}
		for name, body := range assets {
			sum := sha256.Sum256(body)
			lines = append(lines, fmt.Sprintf("%s  %s", hex.EncodeToString(sum[:]), name))
		}
		assets["checksums.txt"] = []byte(strings.Join(lines, "\n") + "\n")
	}
	g.releases[tag] = &fakeRelease{tag: tag, assets: assets}
	g.latest = tag
}

func (g *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.requests = append(g.requests, r.URL.Path)

	writeRelease := func(rel *fakeRelease) {
		out := release{Tag: rel.tag}
		for name := range rel.assets {
			out.Assets = append(out.Assets, releaseAsset{
				Name:   name,
				APIURL: "https://api.github.com/assets/" + rel.tag + "/" + name,
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": out.Tag, "assets": out.Assets})
	}

	switch {
	case strings.HasSuffix(r.URL.Path, "/releases/latest"):
		if g.latest == "" {
			http.NotFound(w, r)
			return
		}
		writeRelease(g.releases[g.latest])
	case strings.Contains(r.URL.Path, "/releases/tags/"):
		tag := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		rel, ok := g.releases[tag]
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeRelease(rel)
	case strings.HasPrefix(r.URL.Path, "/assets/"):
		parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/assets/"), "/", 2)
		rel, ok := g.releases[parts[0]]
		if !ok {
			http.NotFound(w, r)
			return
		}
		body, ok := rel.assets[parts[1]]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	default:
		http.NotFound(w, r)
	}
}

// redirectTransport sends all requests to the test server
type redirectTransport struct {
	target *url.URL
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = t.target.Scheme
	clone.URL.Host = t.target.Host
	return http.DefaultTransport.RoundTrip(clone)
}

type testManager struct {
	*Manager
	dataDir string
	github  *fakeGitHub
}

func newTestManager(t *testing.T, goos, goarch string) *testManager {
	t.Helper()
	gh := newFakeGitHub()
	srv := httptest.NewServer(gh)
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)

	dataDir := filepath.Join(t.TempDir(), "extensions")
	ios, _, _, _ := iostreams.Test()
	m := &Manager{
		dataDir: func() string { return dataDir },
		lookPath: func(s string) (string, error) {
			return "", fmt.Errorf("%s not available in tests: %w", s, exec.ErrNotFound)
		},
		newCommand:  exec.Command,
		platform:    func() (string, string) { return goos, goarch },
		rosetta:     func() bool { return false },
		codesign:    func(string) error { return nil },
		client:      &http.Client{Transport: &redirectTransport{target: target}},
		io:          ios,
		allowBinary: true,
	}
	return &testManager{Manager: m, dataDir: dataDir, github: gh}
}

func extensionArchive(t *testing.T, name string, version string) []byte {
	t.Helper()
	path := makeTarGz(t, []archiveEntry{
		{Name: name + "/extension.yaml", Body: "version: v1\nbinary: {}\ncommands:\n  - name: hello\n"},
		{Name: name + "/bin/" + name, Body: "#!/bin/sh\necho " + version + "\n", Mode: 0755},
		{Name: name + "/views/foo.json", Body: "{}"},
		{Name: name + "/templates/foo.jsonnet", Body: "{}"},
	})
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return b
}

func readBinManifest(t *testing.T, dir string) binManifest {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, manifestName))
	require.NoError(t, err)
	var m binManifest
	require.NoError(t, yaml.Unmarshal(b, &m))
	return m
}

func findExtension(t *testing.T, m *Manager, name string) extensions.Extension {
	t.Helper()
	for _, ext := range m.List() {
		if ext.Name() == name {
			return ext
		}
	}
	t.Fatalf("extension %s not found", name)
	return nil
}

func TestInstallBinaryArchive(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	m.github.addRelease("v1.2.0", map[string][]byte{
		"c8y-foo_1.2.0_linux_amd64.tar.gz":  extensionArchive(t, "c8y-foo", "1.2.0"),
		"c8y-foo_1.2.0_darwin_arm64.tar.gz": []byte("other platform"),
	}, true)

	repo := ghrepo.New("owner", "c8y-foo")
	require.NoError(t, m.Install(repo, "c8y-foo", ""))

	dir := filepath.Join(m.dataDir, "c8y-foo")
	manifest := readBinManifest(t, dir)
	assert.Equal(t, "v1.2.0", manifest.Tag)
	assert.Equal(t, "c8y-foo_1.2.0_linux_amd64.tar.gz", manifest.Asset)
	assert.NotEmpty(t, manifest.SHA256)
	assert.False(t, manifest.IsPinned)
	assert.Equal(t, "bin/c8y-foo", manifest.Path)

	ext := findExtension(t, m.Manager, "foo")
	assert.Equal(t, extensions.KindBinary, ext.Kind())
	assert.Equal(t, "v1.2.0", ext.CurrentVersion())
	assert.Equal(t, dir, ext.Path())
	assert.Equal(t, filepath.Join(dir, "views"), ext.ViewPath())
	assert.Equal(t, filepath.Join(dir, "templates"), ext.TemplatePath())
	assert.FileExists(t, ext.Executable())
	assert.FileExists(t, filepath.Join(dir, "views", "foo.json"))

	// No temporary folders are left behind
	entries, err := os.ReadDir(m.dataDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
}

func TestInstallBinaryPinnedAndUpgrade(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	m.github.addRelease("v1.0.0", map[string][]byte{
		"c8y-foo_1.0.0_linux_amd64.tar.gz": extensionArchive(t, "c8y-foo", "1.0.0"),
	}, true)
	m.github.addRelease("v2.0.0", map[string][]byte{
		"c8y-foo_2.0.0_linux_amd64.tar.gz": extensionArchive(t, "c8y-foo", "2.0.0"),
	}, true)

	repo := ghrepo.New("owner", "c8y-foo")
	require.NoError(t, m.Install(repo, "c8y-foo", "v1.0.0"))

	dir := filepath.Join(m.dataDir, "c8y-foo")
	manifest := readBinManifest(t, dir)
	assert.Equal(t, "v1.0.0", manifest.Tag)
	assert.True(t, manifest.IsPinned)

	err := m.Upgrade("foo", false)
	assert.ErrorIs(t, err, ErrPinnedExtensionUpgrade)

	// Reinstall unpinned, then upgrade
	require.NoError(t, m.Remove("foo"))
	m.github.latest = "v1.0.0"
	require.NoError(t, m.Install(repo, "c8y-foo", ""))
	assert.Equal(t, "v1.0.0", readBinManifest(t, dir).Tag)

	m.github.latest = "v2.0.0"
	require.NoError(t, m.Upgrade("foo", false))
	assert.Equal(t, "v2.0.0", readBinManifest(t, dir).Tag)

	b, err := os.ReadFile(filepath.Join(dir, "bin", "c8y-foo"))
	require.NoError(t, err)
	assert.Contains(t, string(b), "2.0.0")

	assert.ErrorIs(t, m.Upgrade("foo", false), ErrUpToDate)
}

func TestInstallBinaryChecksumMismatchKeepsPreviousVersion(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	m.github.addRelease("v1.0.0", map[string][]byte{
		"c8y-foo_1.0.0_linux_amd64.tar.gz": extensionArchive(t, "c8y-foo", "1.0.0"),
	}, true)
	repo := ghrepo.New("owner", "c8y-foo")
	require.NoError(t, m.Install(repo, "c8y-foo", ""))

	m.github.addRelease("v2.0.0", map[string][]byte{
		"c8y-foo_2.0.0_linux_amd64.tar.gz": extensionArchive(t, "c8y-foo", "2.0.0"),
		"checksums.txt":                    []byte("0000  c8y-foo_2.0.0_linux_amd64.tar.gz\n"),
	}, false)

	err := m.Upgrade("foo", false)
	assert.ErrorIs(t, err, ErrChecksumMismatch)

	dir := filepath.Join(m.dataDir, "c8y-foo")
	assert.Equal(t, "v1.0.0", readBinManifest(t, dir).Tag)
	assert.FileExists(t, filepath.Join(dir, "bin", "c8y-foo"))
	entries, _ := os.ReadDir(m.dataDir)
	assert.Len(t, entries, 1)
}

func TestInstallBinaryWithoutChecksums(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	m.github.addRelease("v1.0.0", map[string][]byte{
		"c8y-foo_1.0.0_linux_amd64.tar.gz": extensionArchive(t, "c8y-foo", "1.0.0"),
	}, false)
	require.NoError(t, m.Install(ghrepo.New("owner", "c8y-foo"), "c8y-foo", ""))
	assert.Empty(t, readBinManifest(t, filepath.Join(m.dataDir, "c8y-foo")).SHA256)
	assert.Contains(t, m.io.ErrOut.(fmt.Stringer).String(), "not verified")
}

func TestInstallBinaryMissingPlatform(t *testing.T) {
	m := newTestManager(t, "freebsd", "amd64")
	m.github.addRelease("v1.0.0", map[string][]byte{
		"c8y-foo_1.0.0_linux_amd64.tar.gz": extensionArchive(t, "c8y-foo", "1.0.0"),
	}, true)
	err := m.Install(ghrepo.New("owner", "c8y-foo"), "c8y-foo", "")
	assert.ErrorIs(t, err, ErrNoPlatformAsset)
	assert.ErrorContains(t, err, "owner/c8y-foo")
	assert.ErrorContains(t, err, "freebsd/amd64")
	assert.NoDirExists(t, filepath.Join(m.dataDir, "c8y-foo"))
}

func TestInstallBinaryArchiveWithoutManifest(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	archive := makeTarGz(t, []archiveEntry{{Name: "bin/c8y-foo", Body: "x", Mode: 0755}})
	b, _ := os.ReadFile(archive)
	m.github.addRelease("v1.0.0", map[string][]byte{"c8y-foo_linux_amd64.tar.gz": b}, true)

	err := m.Install(ghrepo.New("owner", "c8y-foo"), "c8y-foo", "")
	assert.ErrorContains(t, err, "extension.yaml")
	entries, _ := os.ReadDir(m.dataDir)
	assert.Empty(t, entries)
}

func TestInstallBareBinary(t *testing.T) {
	m := newTestManager(t, "windows", "amd64")
	m.github.addRelease("v1.0.0", map[string][]byte{
		"c8y-foo-windows-amd64.exe": []byte("MZ binary"),
		"c8y-foo-linux-amd64":       []byte("ELF binary"),
	}, true)
	require.NoError(t, m.Install(ghrepo.New("owner", "c8y-foo"), "c8y-foo", ""))

	dir := filepath.Join(m.dataDir, "c8y-foo")
	assert.FileExists(t, filepath.Join(dir, "c8y-foo.exe"))
	ext := &Extension{path: dir}
	manifest, err := ext.Manifest()
	require.NoError(t, err)
	require.NotNil(t, manifest.Binary)
	assert.Equal(t, "c8y-foo.exe", manifest.Binary.Path)
}

func TestInstallBinaryDryRun(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	m.github.addRelease("v1.0.0", map[string][]byte{
		"c8y-foo_1.0.0_linux_amd64.tar.gz": extensionArchive(t, "c8y-foo", "1.0.0"),
	}, true)
	m.EnableDryRunMode()
	require.NoError(t, m.Install(ghrepo.New("owner", "c8y-foo"), "c8y-foo", ""))
	out := m.io.Out.(fmt.Stringer).String()
	assert.Contains(t, out, "c8y-foo_1.0.0_linux_amd64.tar.gz")
	assert.Contains(t, out, "v1.0.0")
	assert.NoDirExists(t, m.dataDir)

	for _, path := range m.github.requests {
		assert.False(t, strings.HasPrefix(path, "/assets/"), "asset should not be downloaded: %s", path)
	}
}

func TestInstallWithoutReleaseUsesGit(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	err := m.Install(ghrepo.New("owner", "c8y-foo"), "c8y-foo", "")
	// git is not available in the test manager, so this shows git was used
	assert.True(t, errors.Is(err, exec.ErrNotFound), "got %v", err)
}

func TestInstallReleaseWithoutBinariesUsesGit(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	m.github.addRelease("v1.0.0", map[string][]byte{"notes.md": []byte("x")}, false)
	err := m.Install(ghrepo.New("owner", "c8y-foo"), "c8y-foo", "")
	assert.True(t, errors.Is(err, exec.ErrNotFound), "got %v", err)
}

func TestUpgradeMigratesGitExtensionToBinary(t *testing.T) {
	gitExe, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not installed")
	}
	m := newTestManager(t, "linux", "amd64")
	m.github.addRelease("v1.0.0", map[string][]byte{
		"c8y-foo_1.0.0_linux_amd64.tar.gz": extensionArchive(t, "c8y-foo", "1.0.0"),
	}, true)

	dir := filepath.Join(m.dataDir, "c8y-foo")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "commands"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "commands", "old"), []byte("#!/bin/sh\n"), 0755))
	for _, args := range [][]string{
		{"init", "--quiet", dir},
		{"-C", dir, "remote", "add", "origin", "https://github.com/owner/c8y-foo.git"},
	} {
		out, err := exec.Command(gitExe, args...).CombinedOutput()
		require.NoError(t, err, string(out))
	}

	ext := Extension{path: dir, kind: GitKind, currentVersion: "aaaa", latestVersion: "bbbb"}
	require.NoError(t, m.upgradeExtension(ext, false))

	assert.Equal(t, "v1.0.0", readBinManifest(t, dir).Tag)
	assert.NoFileExists(t, filepath.Join(dir, "commands", "old"))
	assert.NoDirExists(t, filepath.Join(dir, ".git"))
	assert.Equal(t, extensions.KindBinary, findExtension(t, m.Manager, "foo").Kind())
}

func TestListIgnoresHiddenFolders(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	require.NoError(t, os.MkdirAll(filepath.Join(m.dataDir, ".tmp-install-c8y-foo"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(m.dataDir, "c8y-bar"), 0755))
	exts := m.List()
	require.Len(t, exts, 1)
	assert.Equal(t, "bar", exts[0].Name())
}

func writeSystemExtension(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	writeTestFile(t, filepath.Join(path, "extension.yaml"), "version: v1\nbinary: {}\n")
	writeTestFile(t, filepath.Join(path, "bin", name), "#!/bin/sh\n")
	writeTestFile(t, filepath.Join(path, "views", "foo.json"), "{}")
	return path
}

func TestSystemExtensions(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	sys1 := filepath.Join(t.TempDir(), "share", "c8y", "extensions")
	sys2 := filepath.Join(t.TempDir(), "other")
	m.systemDirs = func() []string { return []string{sys1, filepath.Join(t.TempDir(), "missing"), sys2} }

	fooPath := writeSystemExtension(t, sys1, "c8y-foo")
	writeSystemExtension(t, sys2, "c8y-foo") // shadowed by the first system folder
	writeSystemExtension(t, sys2, "c8y-bar") // shadowed by the user folder
	writeSystemExtension(t, sys2, "c8y-baz")
	writeTestFile(t, filepath.Join(sys2, "c8y-baz", manifestName), "owner: acme\nname: c8y-baz\nhost: github.com\ntag: v1.2.3\n")
	require.NoError(t, os.MkdirAll(filepath.Join(sys2, ".hidden"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sys2, "not-an-extension"), 0755))

	// user extension (the user data dir does not need to exist for system extensions,
	// but here it overrides c8y-bar)
	require.NoError(t, os.MkdirAll(filepath.Join(m.dataDir, "c8y-bar"), 0755))

	exts := m.List()
	byName := map[string]extensions.Extension{}
	for _, ext := range exts {
		byName[ext.Name()] = ext
	}
	require.Len(t, exts, 3)

	assert.Equal(t, extensions.KindSystem, byName["foo"].Kind())
	assert.Equal(t, fooPath, byName["foo"].Path())
	assert.Equal(t, filepath.Join(fooPath, "bin", "c8y-foo"), byName["foo"].Executable())
	assert.Equal(t, filepath.Join(fooPath, "views"), byName["foo"].ViewPath())

	assert.NotEqual(t, extensions.KindSystem, byName["bar"].Kind())

	assert.Equal(t, "v1.2.3", byName["baz"].CurrentVersion())
	assert.False(t, byName["baz"].UpdateAvailable())

	// read-only
	assert.ErrorIs(t, m.Upgrade("foo", false), ErrSystemExtensionUpgrade)
	err := m.Remove("foo")
	assert.ErrorContains(t, err, "package manager")
	assert.DirExists(t, fooPath)
}

func TestSystemExtensionsWithoutUserDataDir(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	sys := t.TempDir()
	m.systemDirs = func() []string { return []string{sys} }
	writeSystemExtension(t, sys, "c8y-foo")

	exts := m.List()
	require.Len(t, exts, 1)
	assert.Equal(t, "foo", exts[0].Name())
}

func TestSystemExtensionsSymlinkedFolder(t *testing.T) {
	if os.PathSeparator != '/' {
		t.Skip("symlinks require extra permissions on Windows")
	}
	m := newTestManager(t, "linux", "amd64")
	sys := t.TempDir()
	m.systemDirs = func() []string { return []string{sys} }

	// e.g. Homebrew links share/c8y/extensions/c8y-foo to the Cellar
	cellar := writeSystemExtension(t, t.TempDir(), "c8y-foo")
	require.NoError(t, os.Symlink(cellar, filepath.Join(sys, "c8y-foo")))

	exts := m.List()
	require.Len(t, exts, 1)
	assert.Equal(t, extensions.KindSystem, exts[0].Kind())
	assert.FileExists(t, exts[0].Executable())
}

func TestInstallLocalAddsPrefix(t *testing.T) {
	if os.PathSeparator != '/' {
		t.Skip("symlinks require extra permissions on Windows")
	}
	m := newTestManager(t, "linux", "amd64")
	src := writeSystemExtension(t, t.TempDir(), "c8y-foo")
	require.NoError(t, m.InstallLocal(src, "custom"))

	ext := findExtension(t, m.Manager, "custom")
	assert.Equal(t, extensions.KindLocal, ext.Kind())
	assert.FileExists(t, ext.Executable())
}

func TestInstallBinaryWithCustomName(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	m.github.addRelease("v1.0.0", map[string][]byte{
		// extension.yaml uses the default binary path (bin/c8y-foo)
		"c8y-foo_1.0.0_linux_amd64.tar.gz": extensionArchive(t, "c8y-foo", "1.0.0"),
	}, true)
	require.NoError(t, m.Install(ghrepo.New("owner", "c8y-foo"), "c8y-renamed", ""))

	ext := findExtension(t, m.Manager, "renamed")
	assert.Equal(t, filepath.Join(m.dataDir, "c8y-renamed", "bin", "c8y-foo"), ext.Executable())
	assert.FileExists(t, ext.Executable())
}
