package format

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mvdan.cc/sh/v3/syntax"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatScript_BasicFunctionality(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standalone command should add -n",
			input:    "c8y devices list",
			expected: "c8y -n devices list",
		},
		{
			name:     "command feeding a pipeline should use -n",
			input:    "c8y devices list | head -5",
			expected: "c8y -n devices list | head -5",
		},
		{
			name:     "command feeding a pipeline but uses -n should preserve the -n at the existing location",
			input:    "c8y devices list -n | head -5",
			expected: "c8y devices list -n | head -5",
		},
		{
			name:     "multiple standalone commands",
			input:    "c8y devices list\nc8y devices get --id 123",
			expected: "c8y -n devices list\nc8y -n devices get --id 123",
		},
		{
			name:     "mixed pipeline and standalone",
			input:    "c8y devices list -n | c8y devices get --id 123\nc8y alarms list",
			expected: "c8y devices list -n | c8y devices get --id 123\nc8y -n alarms list",
		},
		{
			name:     "command with existing -n in pipeline should remove it",
			input:    "echo 'test' | c8y devices list -n",
			expected: "echo 'test' | c8y devices list",
		},
		{
			name:     "non-c8y commands should be unchanged",
			input:    "ls -la | grep test | head -10",
			expected: "ls -la | grep test | head -10",
		},
		{
			name:     "c8y command with complex arguments",
			input:    "c8y devices list --type 'myType' --fragmentType 'c8y_IsDevice'",
			expected: "c8y -n devices list --type 'myType' --fragmentType 'c8y_IsDevice'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatScript(tt.input, FormatOptions{})
			require.NoError(t, err)
			assert.Equal(t, tt.expected, strings.TrimSuffix(result, "\n"))
		})
	}
}

func TestFormatScript_PipelineDetection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple pipeline where the feeding command requires -n but preserves existing usage and strips -n from commands in the pipeline",
			input:    "c8y devices list -n | c8y devices get -n --id 123",
			expected: "c8y devices list -n | c8y devices get --id 123",
		},
		{
			name:     "pipeline with multiple commands",
			input:    "c8y devices list -n | grep 'device' | c8y devices get -n --id 123 | jq '.name'",
			expected: "c8y devices list -n | grep 'device' | c8y devices get --id 123 | jq '.name'",
		},
		{
			name:     "nested pipelines with subshells",
			input:    "(c8y devices list -n | head -1) | c8y devices get -n --id 123",
			expected: "(c8y devices list -n | head -1) | c8y devices get --id 123",
		},
		{
			name:     "pipeline with redirects",
			input:    "c8y devices list -n 2>/dev/null | c8y devices get -n --id 123",
			expected: "c8y devices list -n 2>/dev/null | c8y devices get --id 123",
		},
		{
			name:     "command with input redirect",
			input:    "c8y devices create < input.json",
			expected: "c8y devices create <input.json",
		},
		{
			name:     "pipeline with input redirect",
			input:    "cat devices.txt | c8y devices get -n --id 123",
			expected: "cat devices.txt | c8y devices get --id 123",
		},
		{
			name:     "complex pipeline with conditionals",
			input:    "c8y devices list -n | { c8y devices get -n --id 123 || echo 'failed'; }",
			expected: "c8y devices list -n | { c8y devices get --id 123 || echo 'failed'; }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatScript(tt.input, FormatOptions{})
			require.NoError(t, err)
			assert.Equal(t, tt.expected, strings.TrimSuffix(result, "\n"))
		})
	}
}

func TestFormatScript_ComplexShellConstructs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "if statement with c8y commands",
			input:    "if c8y devices list -n | grep -q 'device'; then c8y devices get --id 123; fi",
			expected: "if c8y devices list -n | grep -q 'device'; then c8y -n devices get --id 123; fi",
		},
		{
			name:     "for loop",
			input:    "for id in $(c8y devices list -n | jq -r '.id'); do c8y devices get --id $id; done",
			expected: "for id in $(c8y devices list -n | jq -r '.id'); do c8y -n devices get --id $id; done",
		},
		{
			name:     "while loop",
			input:    "while read id; do c8y devices get --id $id; done < <(c8y devices list -n)",
			expected: "while read id; do c8y -n devices get --id $id; done < <(c8y devices list -n)",
		},
		{
			name:     "function definition",
			input:    "get_device() { c8y devices get --id $1; }\nget_device 123",
			expected: "get_device() { c8y -n devices get --id $1; }\nget_device 123",
		},
		{
			name:     "case statement",
			input:    "case $action in create) c8y devices create --name test ;; list) c8y devices list -n ;; esac",
			expected: "case $action in create) c8y -n devices create --name test ;; list) c8y devices list -n ;; esac",
		},
		{
			name:     "command substitution",
			input:    "c8y devices get --id $(c8y devices list -n | head -1 | jq -r '.id')",
			expected: "c8y -n devices get --id $(c8y devices list -n | head -1 | jq -r '.id')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatScript(tt.input, FormatOptions{})
			require.NoError(t, err)
			assert.Equal(t, tt.expected, strings.TrimSuffix(result, "\n"))
		})
	}
}

func TestFormatScript_IdempotentBehavior(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "already formatted script should remain unchanged",
			input: "c8y -n devices list\nc8y -n devices list | c8y devices get --id 123",
		},
		{
			name:  "complex script with mixed states",
			input: "c8y devices list -n | c8y devices get --id 123\nc8y -n alarms list\nc8y -n devices create | c8y devices update --id 456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First pass
			result1, err := FormatScript(tt.input, FormatOptions{})
			require.NoError(t, err)

			// Second pass - should be identical
			result2, err := FormatScript(result1, FormatOptions{})
			require.NoError(t, err)

			assert.Equal(t, result1, result2, "Second formatting pass should produce identical output")
		})
	}
}

func TestFormatScript_CommentAddition(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		addComment bool
		expected   string
	}{
		{
			name:       "add comment to script without shebang",
			input:      "c8y devices list",
			addComment: true,
			expected: `# -----------------------------------------------------------------------------------------------------
# NOTE: This script has been processed by go-c8y-cli
#
# The following modifications were made:
# - Commands in pipelines have -n flags removed as this would prevent the command receiving piped input
# - Commands not in pipelines have -n flags added to prevent issues in CI/CD environments or in
#   'while read' loops
# -----------------------------------------------------------------------------------------------------

c8y -n devices list`,
		},
		{
			name:       "add comment to script with shebang",
			input:      "#!/bin/bash\nc8y devices list",
			addComment: true,
			expected: `#!/bin/bash
# -----------------------------------------------------------------------------------------------------
# NOTE: This script has been processed by go-c8y-cli
#
# The following modifications were made:
# - Commands in pipelines have -n flags removed as this would prevent the command receiving piped input
# - Commands not in pipelines have -n flags added to prevent issues in CI/CD environments or in
#   'while read' loops
# -----------------------------------------------------------------------------------------------------

c8y -n devices list`,
		},
		{
			name:       "don't add duplicate comment",
			input:      "# NOTE: This script has been processed by go-c8y-cli\nc8y devices list",
			addComment: true,
			expected:   "# NOTE: This script has been processed by go-c8y-cli\nc8y -n devices list",
		},
		{
			name:       "no comment when disabled",
			input:      "c8y devices list",
			addComment: false,
			expected:   "c8y -n devices list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatScript(tt.input, FormatOptions{AddComment: tt.addComment})
			require.NoError(t, err)
			assert.Equal(t, tt.expected, strings.TrimSuffix(result, "\n"))
		})
	}
}

func TestFormatScript_Indentation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		indent   int
		expected string
	}{
		{
			name:     "no indentation (tabs)",
			input:    "if true; then\nc8y devices list\nfi",
			indent:   0,
			expected: "if true; then\n\tc8y -n devices list\nfi",
		},
		{
			name:     "2 space indentation",
			input:    "if true; then\nc8y devices list\nfi",
			indent:   2,
			expected: "if true; then\n  c8y -n devices list\nfi",
		},
		{
			name:     "4 space indentation",
			input:    "if true; then\nc8y devices list\nfi",
			indent:   4,
			expected: "if true; then\n    c8y -n devices list\nfi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatScript(tt.input, FormatOptions{Indent: tt.indent})
			require.NoError(t, err)
			assert.Equal(t, tt.expected, strings.TrimSuffix(result, "\n"))
		})
	}
}

func TestFormatScript_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid shell script",
			input:       "c8y devices list",
			expectError: false,
		},
		{
			name:        "invalid shell syntax",
			input:       "c8y devices list |",
			expectError: true,
		},
		{
			name:        "empty input",
			input:       "",
			expectError: false,
		},
		{
			name:        "only whitespace",
			input:       "   \n\t  ",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FormatScript(tt.input, FormatOptions{})
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLitWord(t *testing.T) {
	word := litWord("-n")
	assert.NotNil(t, word)
	assert.Len(t, word.Parts, 1)
	if lit, ok := word.Parts[0].(*syntax.Lit); ok {
		assert.Equal(t, "-n", lit.Value)
	} else {
		t.Errorf("Expected *syntax.Lit, got %T", word.Parts[0])
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copy_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	content := "test content for copying"
	err = os.WriteFile(srcFile, []byte(content), 0644)
	require.NoError(t, err)

	err = copyFile(srcFile, dstFile)
	require.NoError(t, err)

	copiedContent, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(copiedContent))
}
