package cmdparser

import (
	"slices"
	"testing"
)

func Test_parseValues(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   []string
	}{
		{"ids", []string{"12345, 1231123"}, []string{"12345", "1231123"}},
		{"ids and names", []string{"12345, 1231123", "0124,my values"}, []string{"12345", "1231123", "0124", "my values"}},
		{"ignores empty values and pipeline references", []string{"", "-", "-.id", "1,,2"}, []string{"1", "2"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseValues(tc.values); !slices.Equal(got, tc.want) {
				t.Errorf("parseValues() = %v, want %v", got, tc.want)
			}
		})
	}
}
