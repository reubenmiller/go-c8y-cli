package extension

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/ghrepo"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// namedArchive returns a release archive whose extension.yaml has the given content. The
// binary is at the default path (bin/<repository name>)
func namedArchive(t *testing.T, repoName string, manifest string) []byte {
	t.Helper()
	path := makeTarGz(t, []archiveEntry{
		{Name: repoName + "/extension.yaml", Body: manifest},
		{Name: repoName + "/bin/" + repoName, Body: "#!/bin/sh\necho ok\n", Mode: 0755},
	})
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return b
}

// installDirEntries returns the names of the entries in the extension directory
func installDirEntries(t *testing.T, m *testManager) []string {
	t.Helper()
	entries, err := os.ReadDir(m.dataDir)
	if os.IsNotExist(err) {
		return nil
	}
	require.NoError(t, err)
	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestInstallReleaseWithDeclaredName(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	repoName := "c8y-cli-extension-analytics"
	m.github.addRelease("v1.0.0", map[string][]byte{
		repoName + "_1.0.0_linux_amd64.tar.gz": namedArchive(t, repoName, "version: v1\nname: Analytics\nbinary: {}\n"),
	}, true)

	checked := []string{}
	name, err := m.Install(ghrepo.New("owner", repoName), extensions.InstallOptions{
		CheckName: func(n string) error { checked = append(checked, n); return nil },
	})
	require.NoError(t, err)
	assert.Equal(t, "analytics", name)
	assert.Equal(t, []string{"analytics"}, checked)
	assert.Equal(t, []string{"c8y-analytics"}, installDirEntries(t, m))

	ext := findExtension(t, m.Manager, "analytics")
	assert.Equal(t, extensions.KindBinary, ext.Kind())
	// The default binary path and updates use the repository name
	assert.Equal(t, repoName, readBinManifest(t, ext.Path()).Name)
	assert.FileExists(t, ext.Executable())
}

func TestInstallReleaseNameFlagOverridesDeclaredName(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	repoName := "c8y-cli-extension-analytics"
	m.github.addRelease("v1.0.0", map[string][]byte{
		repoName + "_1.0.0_linux_amd64.tar.gz": namedArchive(t, repoName, "version: v1\nname: analytics\nbinary: {}\n"),
	}, true)

	name, err := m.Install(ghrepo.New("owner", repoName), extensions.InstallOptions{Name: "myanalytics"})
	require.NoError(t, err)
	assert.Equal(t, "myanalytics", name)
	assert.Equal(t, []string{"c8y-myanalytics"}, installDirEntries(t, m))
}

func TestInstallReleaseWithoutDeclaredName(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	m.github.addRelease("v1.0.0", map[string][]byte{
		"c8y-foo_1.0.0_linux_amd64.tar.gz": namedArchive(t, "c8y-foo", "version: v1\nbinary: {}\n"),
	}, true)

	name, err := m.Install(ghrepo.New("owner", "c8y-foo"), extensions.InstallOptions{})
	require.NoError(t, err)
	assert.Equal(t, "foo", name)
	assert.Equal(t, []string{"c8y-foo"}, installDirEntries(t, m))
}

func TestInstallReleaseFailuresLeaveNothing(t *testing.T) {
	tests := []struct {
		name      string
		repoName  string
		manifest  string
		checkName func(string) error
		wantErr   string
		errIs     error
	}{
		{
			name:     "invalid declared name",
			repoName: "c8y-foo",
			manifest: "version: v1\nname: My Tool\nbinary: {}\n",
			wantErr:  `invalid extension name "My Tool"`,
		},
		{
			name:      "name check fails",
			repoName:  "c8y-foo",
			manifest:  "version: v1\nname: devices\nbinary: {}\n",
			checkName: func(n string) error { return errors.New(n + " is a built-in command") },
			wantErr:   "devices is a built-in command",
		},
		{
			name:     "repository without the prefix and without a name",
			repoName: "analytics-tool",
			manifest: "version: v1\nbinary: {}\n",
			errIs:    ErrInvalidRepositoryName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestManager(t, "linux", "amd64")
			m.github.addRelease("v1.0.0", map[string][]byte{
				tt.repoName + "_1.0.0_linux_amd64.tar.gz": namedArchive(t, tt.repoName, tt.manifest),
			}, true)

			_, err := m.Install(ghrepo.New("owner", tt.repoName), extensions.InstallOptions{CheckName: tt.checkName})
			require.Error(t, err)
			if tt.errIs != nil {
				assert.ErrorIs(t, err, tt.errIs)
			}
			if tt.wantErr != "" {
				assert.Contains(t, err.Error(), tt.wantErr)
			}
			assert.Empty(t, installDirEntries(t, m), "no extension or temporary folder is left")
		})
	}
}

func TestInstallReleaseRepositoryWithoutPrefixWithDeclaredName(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	m.github.addRelease("v1.0.0", map[string][]byte{
		"analytics-tool_1.0.0_linux_amd64.tar.gz": namedArchive(t, "analytics-tool", "version: v1\nname: analytics\nbinary: {}\n"),
	}, true)

	name, err := m.Install(ghrepo.New("owner", "analytics-tool"), extensions.InstallOptions{})
	require.NoError(t, err)
	assert.Equal(t, "analytics", name)
	assert.Equal(t, []string{"c8y-analytics"}, installDirEntries(t, m))
}

func TestUpdateKeepsInstalledFolder(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	repoName := "c8y-cli-extension-analytics"
	m.github.addRelease("v1.0.0", map[string][]byte{
		repoName + "_1.0.0_linux_amd64.tar.gz": namedArchive(t, repoName, "version: v1\nname: analytics\nbinary: {}\n"),
	}, true)
	_, err := m.Install(ghrepo.New("owner", repoName), extensions.InstallOptions{})
	require.NoError(t, err)

	m.github.addRelease("v2.0.0", map[string][]byte{
		repoName + "_2.0.0_linux_amd64.tar.gz": namedArchive(t, repoName, "version: v1\nname: dtm-analytics\nbinary: {}\n"),
	}, true)
	require.NoError(t, m.installBin(ghrepo.New("owner", repoName), "c8y-analytics", ""))

	assert.Equal(t, []string{"c8y-analytics"}, installDirEntries(t, m))
	assert.Equal(t, "v2.0.0", readBinManifest(t, filepath.Join(m.dataDir, "c8y-analytics")).Tag)
}

// useFakeGit makes the manager use a fake git, whose clone copies src. Calls are logged to the returned file
func useFakeGit(t *testing.T, m *testManager, src string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake git is a shell script")
	}
	dir := t.TempDir()
	logFile := filepath.Join(dir, "git.log")
	script := filepath.Join(dir, "git")
	writeTestFile(t, script, `#!/bin/sh
echo "$@" >> "`+logFile+`"
if [ "$1" = "clone" ]; then
  cp -R "`+src+`" "$3"
fi
`)
	require.NoError(t, os.Chmod(script, 0755))
	m.lookPath = func(string) (string, error) { return script, nil }
	return logFile
}

func TestInstallGitWithDeclaredName(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	src := t.TempDir()
	writeTestFile(t, filepath.Join(src, "extension.yaml"), "version: v1\nname: analytics\n")
	writeTestFile(t, filepath.Join(src, "commands", "hello"), "#!/bin/sh\necho hello\n")
	logFile := useFakeGit(t, m, src)

	checked := []string{}
	name, err := m.Install(ghrepo.New("owner", "c8y-cli-extension-analytics"), extensions.InstallOptions{
		Pin:       "v1.0.0",
		CheckName: func(n string) error { checked = append(checked, n); return nil },
	})
	require.NoError(t, err)
	assert.Equal(t, "analytics", name)
	assert.Equal(t, []string{"analytics"}, checked)
	assert.Equal(t, []string{"c8y-analytics"}, installDirEntries(t, m))
	assert.FileExists(t, filepath.Join(m.dataDir, "c8y-analytics", ".pin-0123456789abcdef0123456789abcdef01234567"))

	log, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.Contains(t, string(log), "checkout 0123456789abcdef0123456789abcdef01234567")

	// IsPinned compares the pin file with "git rev-parse", which the fake git doesn't support
	assert.Equal(t, extensions.KindGit, findExtension(t, m.Manager, "analytics").Kind())
}

func TestInstallGitNameCheckFailsLeavesNothing(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	src := t.TempDir()
	writeTestFile(t, filepath.Join(src, "extension.yaml"), "version: v1\nname: devices\n")
	useFakeGit(t, m, src)

	_, err := m.Install(ghrepo.New("owner", "c8y-tools"), extensions.InstallOptions{
		CheckName: func(n string) error { return errors.New("clash") },
	})
	require.EqualError(t, err, "clash")
	assert.Empty(t, installDirEntries(t, m))
}

func TestInstallGitDryRun(t *testing.T) {
	m := newTestManager(t, "linux", "amd64")
	logFile := useFakeGit(t, m, t.TempDir())
	m.dryRunMode = true

	name, err := m.Install(ghrepo.New("owner", "c8y-tools"), extensions.InstallOptions{})
	require.NoError(t, err)
	assert.Empty(t, name)
	assert.NoFileExists(t, logFile, "git is not run")
	assert.Empty(t, installDirEntries(t, m))
}

func TestInstallLocalWithDeclaredName(t *testing.T) {
	if os.PathSeparator != '/' {
		t.Skip("symlinks require extra permissions on Windows")
	}
	m := newTestManager(t, "linux", "amd64")
	src := filepath.Join(t.TempDir(), "my-analytics-tool")
	writeTestFile(t, filepath.Join(src, "extension.yaml"), "version: v1\nname: analytics\n")

	failCheck := func(n string) error { return errors.New("already installed: " + n) }

	name, err := m.InstallLocal(src, extensions.InstallOptions{CheckName: func(string) error { return nil }})
	require.NoError(t, err)
	assert.Equal(t, "analytics", name)
	assert.Equal(t, []string{"c8y-analytics"}, installDirEntries(t, m))

	// Installing the same folder again doesn't run the name check
	name, err = m.InstallLocal(src, extensions.InstallOptions{CheckName: failCheck})
	require.NoError(t, err)
	assert.Equal(t, "analytics", name)

	// Another folder with the same name is checked
	other := filepath.Join(t.TempDir(), "other")
	writeTestFile(t, filepath.Join(other, "extension.yaml"), "version: v1\nname: analytics\n")
	_, err = m.InstallLocal(other, extensions.InstallOptions{CheckName: failCheck})
	require.EqualError(t, err, "already installed: analytics")

	ext := findExtension(t, m.Manager, "analytics")
	assert.Equal(t, extensions.KindLocal, ext.Kind())
	target, ok := localLinkTarget(filepath.Join(m.dataDir, "c8y-analytics"))
	require.True(t, ok)
	assert.True(t, sameDir(target, src))
}

func TestInstallLocalWithoutDeclaredName(t *testing.T) {
	if os.PathSeparator != '/' {
		t.Skip("symlinks require extra permissions on Windows")
	}
	m := newTestManager(t, "linux", "amd64")
	src := filepath.Join(t.TempDir(), "tools")
	writeTestFile(t, filepath.Join(src, "extension.yaml"), "version: v1\n")

	name, err := m.InstallLocal(src, extensions.InstallOptions{})
	require.NoError(t, err)
	assert.Equal(t, "tools", name)
	assert.True(t, strings.HasPrefix(installDirEntries(t, m)[0], ExtPrefix))
}

func TestUpdatePrintsNameChange(t *testing.T) {
	repoName := "c8y-cli-extension-analytics"
	tests := []struct {
		name       string
		installAs  string
		v1Name     string
		v2Name     string
		wantNotice bool
	}{
		{"declared name changes", "", "analytics", "dtm-analytics", true},
		{"declared name is unchanged", "", "analytics", "analytics", false},
		{"installed with --name and unchanged", "myanalytics", "analytics", "analytics", false},
		{"installed with --name and changed", "myanalytics", "analytics", "dtm-analytics", true},
		{"changed to the installed name", "dtm-analytics", "analytics", "dtm-analytics", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestManager(t, "linux", "amd64")
			ios, _, _, stderr := iostreams.Test()
			m.io = ios
			m.github.addRelease("v1.0.0", map[string][]byte{
				repoName + "_1.0.0_linux_amd64.tar.gz": namedArchive(t, repoName, "version: v1\nname: "+tt.v1Name+"\nbinary: {}\n"),
			}, true)
			m.github.addRelease("v2.0.0", map[string][]byte{
				repoName + "_2.0.0_linux_amd64.tar.gz": namedArchive(t, repoName, "version: v1\nname: "+tt.v2Name+"\nbinary: {}\n"),
			}, true)

			m.github.latest = "v1.0.0"
			name, err := m.Install(ghrepo.New("owner", repoName), extensions.InstallOptions{Name: tt.installAs})
			require.NoError(t, err)

			m.github.latest = "v2.0.0"
			require.NoError(t, m.Upgrade(name, false))
			assert.Equal(t, []string{ExtPrefix + name}, installDirEntries(t, m), "the extension keeps its folder")

			if tt.wantNotice {
				assert.Contains(t, stderr.String(), "now declares the name '"+tt.v2Name+"'")
				assert.Contains(t, stderr.String(), "c8y extensions delete "+name)
				assert.Contains(t, stderr.String(), "c8y extensions install owner/"+repoName)
			} else {
				assert.NotContains(t, stderr.String(), "now declares the name")
			}
		})
	}
}
