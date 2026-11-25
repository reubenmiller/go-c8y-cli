package login

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// providerScript writes a session provider which returns a session, and records its arguments
// and some environment variables in the session's version field
const providerScript = `#!/bin/sh
if [ -n "$PROVIDER_EXIT" ]; then
  echo "provider failed" >&2
  exit "$PROVIDER_EXIT"
fi
printf '{"host":"https://picked.example.com","tenant":"t123","username":"user1","version":"args=%s|bw=%s|host=%s|token=%s|ext=%s|dir=%s|version=%s"}\n' \
  "$*" "${BW_SESSION:-<unset>}" "${C8Y_HOST:-<unset>}" "${C8Y_TOKEN:-<unset>}" "${C8Y_EXTENSION_NAME:-<unset>}" "${C8Y_EXTENSION_DIR:-<unset>}" "${C8Y_EXTENSION_CONTEXT_VERSION:-<unset>}"
`

func skipOnWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("uses shell scripts as test binaries")
	}
}

func testExtension(t *testing.T, name string, manifest *extensions.Manifest, writeBinary bool) extensions.Extension {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "c8y-"+name)
	exe := filepath.Join(dir, "bin", "c8y-"+name)
	if writeBinary {
		require.NoError(t, os.MkdirAll(filepath.Dir(exe), 0755))
		require.NoError(t, os.WriteFile(exe, []byte(providerScript), 0755))
	}
	return &extensions.ExtensionMock{
		NameFunc:       func() string { return name },
		PathFunc:       func() string { return dir },
		ExecutableFunc: func() string { return exe },
		ManifestFunc:   func() (*extensions.Manifest, error) { return manifest, nil },
	}
}

func bitwardenManifest() *extensions.Manifest {
	return &extensions.Manifest{
		Binary: &extensions.BinarySpec{},
		SessionProviders: []extensions.SessionProvider{
			{Name: "bitwarden", Exec: []string{"list"}, Secrets: []string{"BW_SESSION"}, Format: "json"},
		},
	}
}

func newTestLogin(t *testing.T, v *viper.Viper, exts ...extensions.Extension) *CmdLogin {
	t.Helper()
	if v == nil {
		v = viper.New()
	}
	cfg := config.NewConfig(v)
	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{
		IOStreams:  ios,
		Executable: "c8y",
		Config:     func() (*config.Config, error) { return cfg, nil },
		Logger:     func() (*logger.Logger, error) { return cfg.Logger, nil },
		ExtensionManager: func() extensions.ExtensionManager {
			return &extensions.ExtensionManagerMock{ListFunc: func() []extensions.Extension { return exts }}
		},
	}
	return &CmdLogin{factory: f}
}

func descriptionValues(t *testing.T, description string) map[string]string {
	t.Helper()
	values := map[string]string{}
	for _, part := range strings.Split(description, "|") {
		key, value, _ := strings.Cut(part, "=")
		values[key] = value
	}
	return values
}

func TestExtensionProviderArgumentsAndEnvironment(t *testing.T) {
	skipOnWindows(t)
	t.Setenv("BW_SESSION", "bw-token")
	t.Setenv("C8Y_HOST", "https://active.example.com")
	t.Setenv("C8Y_TOKEN", "active-token")

	ext := testExtension(t, "session-bitwarden", bitwardenManifest(), true)
	n := newTestLogin(t, nil, ext)
	n.ClearToken = true
	n.LoginType = "OAUTH2_INTERNAL"

	session, err := n.FromExtensionProvider("bitwarden", []string{"--folder", "c8y"})
	require.NoError(t, err)
	assert.Equal(t, "https://picked.example.com", session.Host)
	assert.Equal(t, "t123", session.Tenant)

	values := descriptionValues(t, session.Version)
	assert.Equal(t, "list --clear --loginType=OAUTH2_INTERNAL --folder c8y", values["args"])
	assert.Equal(t, "bw-token", values["bw"], "declared secret should be passed")
	assert.Equal(t, "<unset>", values["host"], "active session should not be passed")
	assert.Equal(t, "<unset>", values["token"], "active session should not be passed")
	assert.Equal(t, "session-bitwarden", values["ext"])
	assert.Equal(t, ext.Path(), values["dir"])
	assert.Equal(t, extensions.ContextVersion, values["version"])
}

func TestExtensionProviderQualifiedReference(t *testing.T) {
	skipOnWindows(t)
	ext := testExtension(t, "session-bitwarden", bitwardenManifest(), true)
	n := newTestLogin(t, nil, ext)
	session, err := n.FromExtensionProvider("session-bitwarden/bitwarden", nil)
	require.NoError(t, err)
	assert.Equal(t, "list", descriptionValues(t, session.Version)["args"])
}

func TestExtensionProviderSecretFromPinEntry(t *testing.T) {
	skipOnWindows(t)
	t.Setenv("BW_SESSION", "")
	pinEntry := filepath.Join(t.TempDir(), "pinentry")
	require.NoError(t, os.WriteFile(pinEntry, []byte("#!/bin/sh\necho \"pin-$2\"\n"), 0755))

	v := viper.New()
	v.Set(config.SettingsPinEntry, pinEntry+" get")
	n := newTestLogin(t, v, testExtension(t, "session-bitwarden", bitwardenManifest(), true))

	session, err := n.FromExtensionProvider("bitwarden", nil)
	require.NoError(t, err)
	assert.Equal(t, "pin-BW_SESSION", descriptionValues(t, session.Version)["bw"])
}

func TestExtensionProviderAdditionalSecrets(t *testing.T) {
	skipOnWindows(t)
	// A secret configured by the user is passed even if it is a session variable
	t.Setenv("C8Y_TOKEN", "user-secret")
	n := newTestLogin(t, nil, testExtension(t, "session-bitwarden", bitwardenManifest(), true))
	n.Secrets = []string{"C8Y_TOKEN"}

	session, err := n.FromExtensionProvider("bitwarden", nil)
	require.NoError(t, err)
	assert.Equal(t, "user-secret", descriptionValues(t, session.Version)["token"])
}

func TestExtensionProviderErrors(t *testing.T) {
	skipOnWindows(t)

	t.Run("missing binary", func(t *testing.T) {
		n := newTestLogin(t, nil, testExtension(t, "session-bitwarden", bitwardenManifest(), false))
		_, err := n.FromExtensionProvider("bitwarden", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session-bitwarden")
		assert.Contains(t, err.Error(), filepath.Join("bin", "c8y-session-bitwarden"))
	})

	t.Run("unknown provider", func(t *testing.T) {
		n := newTestLogin(t, nil, testExtension(t, "session-bitwarden", bitwardenManifest(), true))
		_, err := n.FromExtensionProvider("keepass", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session-bitwarden/bitwarden")
	})

	t.Run("no provider set", func(t *testing.T) {
		n := newTestLogin(t, nil)
		_, err := n.FromExtensionProvider("", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session.provider.extension")
	})

	t.Run("provider fails", func(t *testing.T) {
		t.Setenv("PROVIDER_EXIT", "3")
		n := newTestLogin(t, nil, testExtension(t, "session-bitwarden", bitwardenManifest(), true))
		_, err := n.FromExtensionProvider("bitwarden", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exit status 3")
	})
}

func TestExtensionProviderFormat(t *testing.T) {
	skipOnWindows(t)
	manifest := bitwardenManifest()
	manifest.SessionProviders[0].Format = "dotenv"
	n := newTestLogin(t, nil, testExtension(t, "session-bitwarden", manifest, true))

	// The declared format is used: json output parsed as dotenv gives no host
	session, err := n.FromExtensionProvider("bitwarden", nil)
	if err == nil {
		assert.NotEqual(t, "https://picked.example.com", session.Host)
	}

	// The --format flag takes precedence over the declared format
	n.Format = "json"
	session, err = n.FromExtensionProvider("bitwarden", nil)
	require.NoError(t, err)
	assert.Equal(t, "https://picked.example.com", session.Host)
}

func TestExternalProviderUnchanged(t *testing.T) {
	skipOnWindows(t)
	t.Setenv("C8Y_HOST", "https://active.example.com")
	t.Setenv("BW_SESSION", "bw-token")
	script := filepath.Join(t.TempDir(), "provider")
	require.NoError(t, os.WriteFile(script, []byte(providerScript), 0755))

	n := newTestLogin(t, nil)
	n.Exec = script + " list --folder c8y"
	n.ClearToken = true

	session, err := n.FromExternalProvider([]string{"extra"})
	require.NoError(t, err)
	values := descriptionValues(t, session.Version)
	assert.Equal(t, "list --folder c8y --clear extra", values["args"])
	assert.Equal(t, "https://active.example.com", values["host"], "external providers get the full environment")
	assert.Equal(t, "bw-token", values["bw"])
	assert.Equal(t, "<unset>", values["ext"])
}
