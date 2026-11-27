package main

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/clibuild"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
)

// The generated text must not depend on the terminal: even when the process
// reports that stdout is not a terminal, the interactive defaults are used.
func TestPinTerminal(t *testing.T) {
	t.Setenv(iostreams.EnvForceStdoutTTY, "false")
	t.Setenv(iostreams.EnvForceStdinTTY, "false")
	t.Setenv(iostreams.EnvForceStderrTTY, "false")
	t.Setenv("C8Y_SETTINGS_EXTENSIONS_DATADIR", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	pinTerminal()

	rootCmd, err := clibuild.NewRootCommand()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"output": "table", "view": "auto", "compact": "false", "noColor": "false"}
	for name, value := range want {
		f := rootCmd.PersistentFlags().Lookup(name)
		if f == nil {
			t.Fatalf("flag %s not found", name)
		}
		if f.DefValue != value {
			t.Errorf("flag %s default = %q, want %q", name, f.DefValue, value)
		}
	}
}
