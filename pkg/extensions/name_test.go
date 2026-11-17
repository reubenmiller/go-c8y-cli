package extensions

import "testing"

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{"analytics", "analytics", false},
		{"c8y-analytics", "analytics", false},
		{"Analytics", "analytics", false},
		{"C8Y-Device-Simulator", "device-simulator", false},
		{"  analytics ", "analytics", false},
		{"tool2", "tool2", false},
		{"", "", true},
		{"c8y-", "", true},
		{"My Tool", "", true},
		{"my_tool", "", true},
		{"-tool", "", true},
		{"../tool", "", true},
		{"tools/sub", "", true},
	}
	for _, tt := range tests {
		got, err := NormalizeName(tt.name)
		if (err != nil) != tt.wantErr {
			t.Errorf("NormalizeName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
