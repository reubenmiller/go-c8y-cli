package login

import (
	"os"
	"strings"
	"testing"
)

// TestMain runs the tests without the caller's Cumulocity settings: a
// C8Y_SETTINGS_PINENTRY of the developer (e.g. a Touch ID helper) would
// otherwise be run for every secret a test leaves unset.
func TestMain(m *testing.M) {
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(name, "C8Y_") || name == "BW_SESSION" {
			os.Unsetenv(name)
		}
	}
	os.Exit(m.Run())
}
