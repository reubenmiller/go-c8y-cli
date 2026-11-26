package login

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// listingProviderScript prints a json array for "list", and a session for "get <uri>". It
// records whether stdin is a terminal or empty, so tests can check that no stdin is passed
const listingProviderScript = `#!/bin/sh
if [ -n "$PROVIDER_EXIT" ]; then
  echo "debug output" >&2
  echo "vault is locked" >&2
  exit "$PROVIDER_EXIT"
fi
stdin=$(cat)
case "$1" in
  list)
    printf '[{"sessionUri":"vault://1","name":"One","host":"https://one.example.com","password":"leaked","mode":"dev","group":"c8y"},{"sessionUri":"","host":"https://skipped.example.com"},{"sessionUri":"vault://2","host":"https://two.example.com"}]\n'
    ;;
  *)
    printf '{"host":"https://selected.example.com","tenant":"t1","username":"u1","version":"args=%s|stdin=%s|bw=%s"}\n' "$*" "${stdin:-<empty>}" "${BW_SESSION:-<unset>}"
    ;;
esac
`

func listingManifest(list, sel []string) *extensions.Manifest {
	return &extensions.Manifest{
		Binary: &extensions.BinarySpec{},
		SessionProviders: []extensions.SessionProvider{
			{Name: "vault", Exec: []string{"pick"}, List: list, Select: sel, Secrets: []string{"BW_SESSION"}, Format: "json"},
		},
	}
}

func listingExtension(t *testing.T, manifest *extensions.Manifest) extensions.Extension {
	t.Helper()
	ext := testExtension(t, "session-vault", manifest, true)
	require.NoError(t, os.WriteFile(ext.Executable(), []byte(listingProviderScript), 0755))
	return ext
}

func TestExtensionProviderSelect(t *testing.T) {
	skipOnWindows(t)
	t.Setenv("BW_SESSION", "bw-token")
	n := newTestLogin(t, nil, listingExtension(t, listingManifest([]string{"list", "--json"}, []string{"get"})))
	n.Select = "vault://1"
	n.LoginType = "BASIC"

	session, err := n.FromExtensionProvider("vault", []string{"--folder", "c8y"})
	require.NoError(t, err)
	assert.Equal(t, "https://selected.example.com", session.Host)
	values := descriptionValues(t, session.Version)
	assert.Equal(t, "get --loginType=BASIC --folder c8y vault://1", values["args"], "select args, provider flags, user args, then the sessionUri")
	assert.Equal(t, "<empty>", values["stdin"], "the provider must not get stdin")
	assert.Equal(t, "bw-token", values["bw"])
}

func TestExtensionProviderSelectNotSupported(t *testing.T) {
	skipOnWindows(t)
	n := newTestLogin(t, nil, listingExtension(t, listingManifest(nil, nil)))
	n.Select = "vault://1"
	_, err := n.FromExtensionProvider("vault", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support --select")
}

func TestListExtensionSessions(t *testing.T) {
	skipOnWindows(t)
	t.Setenv("BW_SESSION", "bw-token")
	n := newTestLogin(t, nil, listingExtension(t, listingManifest([]string{"list", "--json"}, []string{"get"})))

	provider, sessions, err := ListExtensionSessions(context.Background(), n.factory, "vault", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "session-vault/vault", provider.Ref())
	require.Len(t, sessions, 2, "the session without a sessionUri is skipped")
	assert.Equal(t, extensions.ProviderSession{SessionURI: "vault://1", Name: "One", Host: "https://one.example.com", Mode: "dev", Group: "c8y"}, sessions[0])
	assert.Equal(t, "vault://2", sessions[1].DisplayName())
}

func TestListExtensionSessionsErrors(t *testing.T) {
	skipOnWindows(t)

	t.Run("provider can't list", func(t *testing.T) {
		n := newTestLogin(t, nil, listingExtension(t, listingManifest(nil, nil)))
		_, _, err := ListExtensionSessions(context.Background(), n.factory, "vault", nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can't list its sessions")
	})

	t.Run("provider fails", func(t *testing.T) {
		t.Setenv("PROVIDER_EXIT", "2")
		n := newTestLogin(t, nil, listingExtension(t, listingManifest([]string{"list"}, []string{"get"})))
		_, _, err := ListExtensionSessions(context.Background(), n.factory, "vault", nil, nil)
		require.Error(t, err)
		assert.Equal(t, "vault is locked", err.Error(), "the last line of stderr explains the error")
	})

	t.Run("invalid output", func(t *testing.T) {
		n := newTestLogin(t, nil, listingExtension(t, listingManifest([]string{"get"}, []string{"get"})))
		_, _, err := ListExtensionSessions(context.Background(), n.factory, "vault", nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "json array")
	})

	t.Run("missing binary", func(t *testing.T) {
		ext := testExtension(t, "session-vault", listingManifest([]string{"list"}, []string{"get"}), false)
		n := newTestLogin(t, nil, ext)
		_, _, err := ListExtensionSessions(context.Background(), n.factory, "vault", nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), filepath.Join("bin", "c8y-session-vault"))
	})
}
