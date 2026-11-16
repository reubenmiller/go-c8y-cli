package prompt

import "testing"

func TestSelectWithoutTerminalReturnsDefaultValue(t *testing.T) {
	// Tests don't run in an interactive terminal, so the default option is returned
	options := []string{
		"dev\tDevelopment mode (no restrictions)",
		"prod\tProduction mode (read only)",
	}
	cases := []struct {
		defaultOption string
		want          string
	}{
		{options[0], "dev"},
		{options[1], "prod"},
		{"plain", "plain"},
	}
	for _, tc := range cases {
		got, err := Select("Select mode", options, tc.defaultOption)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if got != tc.want {
			t.Errorf("expected %q, got %q", tc.want, got)
		}
	}
}
