package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestRemoteAccessSSHFallbackUser(t *testing.T) {
	t.Setenv("C8Y_SETTINGS_REMOTEACCESS_SSHFALLBACKUSER", "")
	if got := NewConfig(viper.New()).GetRemoteAccessSSHFallbackUser(); got != "" {
		t.Fatalf("expected the fallback to be disabled by default, got %q", got)
	}

	t.Setenv("C8Y_SETTINGS_REMOTEACCESS_SSHFALLBACKUSER", "root")
	if got := NewConfig(viper.New()).GetRemoteAccessSSHFallbackUser(); got != "root" {
		t.Fatalf("expected the env variable to set the fallback user, got %q", got)
	}
}
