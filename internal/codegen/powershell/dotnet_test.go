package powershell

import (
	"regexp"
	"testing"
)

func TestDotnetReplace(t *testing.T) {
	re := regexp.MustCompile(`\{\{\s*Command\s*\}\}`)

	tests := []struct {
		name        string
		input       string
		replacement string
		want        string
	}{
		{
			name:        "literal dollar names stay literal",
			input:       "run {{ Command }} now",
			replacement: "$TestDevice.id",
			want:        "run $TestDevice.id now",
		},
		{
			name:        "dollar underscore expands to the entire input",
			input:       "a {{ Command }} b",
			replacement: "x$_y",
			want:        "a xa {{ Command }} by b",
		},
		{
			name:        "escaped dollar",
			input:       "{{ Command }}",
			replacement: "a$$b",
			want:        "a$b",
		},
		{
			name:        "whole match and surroundings",
			input:       "a{{ Command }}b",
			replacement: "[$&|$`|$']",
			want:        "a[{{ Command }}|a|b]b",
		},
		{
			name:        "undefined group reference stays literal",
			input:       "{{ Command }}",
			replacement: "keep $1 and ${name}",
			want:        "keep $1 and ${name}",
		},
		{
			name:        "group zero is the whole match",
			input:       "{{ Command }}",
			replacement: "<$0>",
			want:        "<{{ Command }}>",
		},
		{
			name:        "no match returns input",
			input:       "nothing here",
			replacement: "x",
			want:        "nothing here",
		},
		{
			name:        "trailing dollar is literal",
			input:       "{{ Command }}",
			replacement: "cost $",
			want:        "cost $",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dotnetReplace(re, tt.input, tt.replacement); got != tt.want {
				t.Errorf("dotnetReplace() = %q, want %q", got, tt.want)
			}
		})
	}
}
