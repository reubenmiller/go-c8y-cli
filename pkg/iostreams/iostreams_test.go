package iostreams

import "testing"

func Test_SystemForceTTY(t *testing.T) {
	t.Setenv(EnvForceStdinTTY, "true")
	t.Setenv(EnvForceStdoutTTY, "true")
	t.Setenv(EnvForceStderrTTY, "true")

	s := System(true, false)
	if !s.IsStdinTTY() || !s.IsStdoutTTY() || !s.IsStderrTTY() {
		t.Fatalf("expected all streams to be detected as a terminal")
	}
	if !s.progressIndicatorEnabled {
		t.Fatalf("expected progress indicator to be enabled when stderr is a terminal")
	}
}

func Test_SystemForceNoTTY(t *testing.T) {
	t.Setenv(EnvForceStdinTTY, "false")
	t.Setenv(EnvForceStdoutTTY, "false")
	t.Setenv(EnvForceStderrTTY, "false")

	s := System(true, false)
	if s.IsStdinTTY() || s.IsStdoutTTY() || s.IsStderrTTY() {
		t.Fatalf("expected no streams to be detected as a terminal")
	}
	if s.progressIndicatorEnabled {
		t.Fatalf("expected progress indicator to be disabled when stderr is not a terminal")
	}
}
