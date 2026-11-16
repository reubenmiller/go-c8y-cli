package extension

import (
	"testing"

	"golang.org/x/mod/semver"
)

func TestExtSDKVersion(t *testing.T) {
	if !semver.IsValid(defaultExtSDKVersion) {
		t.Fatalf("defaultExtSDKVersion %q is not a valid version", defaultExtSDKVersion)
	}

	orig := extSDKVersion
	t.Cleanup(func() { extSDKVersion = orig })

	tests := []struct {
		value string
		want  string
	}{
		{"", defaultExtSDKVersion},
		{"v0.2.0", "v0.2.0"},
		{"v1.0.0-rc.1", "v1.0.0-rc.1"},
		{"pkg/extsdk/v0.2.0", defaultExtSDKVersion},
		{"0.2.0", defaultExtSDKVersion},
	}
	for _, tt := range tests {
		extSDKVersion = tt.value
		if got := ExtSDKVersion(); got != tt.want {
			t.Errorf("extSDKVersion=%q: got %q, want %q", tt.value, got, tt.want)
		}
	}
}
