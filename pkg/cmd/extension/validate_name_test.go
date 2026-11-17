package extension

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRootCommand() *cobra.Command {
	root := &cobra.Command{Use: "c8y"}
	root.AddCommand(&cobra.Command{Use: "devices", Run: func(*cobra.Command, []string) {}})
	// Command of an installed extension, which is not a built-in command
	root.AddCommand(&cobra.Command{
		Use:         "analytics",
		Annotations: map[string]string{extensions.CommandAnnotation: "analytics"},
		Run:         func(*cobra.Command, []string) {},
	})
	return root
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name      string
		folder    string
		manifest  string
		installed bool
		want      []string
	}{
		{"valid name", "c8y-analytics", "name: analytics\n", false, nil},
		{"no name", "c8y-analytics", "version: v1\n", false, nil},
		{"invalid name", "c8y-tool", "name: My_Tool\n", false, []string{"error|name|extension.yaml|invalid name 'My_Tool'"}},
		{"built-in command", "c8y-devices", "name: devices\n", false, []string{"error|name|extension.yaml|name 'devices' is a built-in command"}},
		{"redundant prefix", "c8y-analytics", "name: c8y-analytics\n", false, []string{"warning|name|extension.yaml|the extension is installed as 'analytics'. Use 'name: analytics'"}},
		{"uppercase letters", "c8y-analytics", "name: Analytics\n", false, []string{"warning|name|extension.yaml|the extension is installed as 'analytics'. Use 'name: analytics'"}},
		{"folder name differs, not installed", "my-analytics-tool", "name: analytics\n", false, nil},
		{"installed under another name", "c8y-myanalytics", "name: analytics\n", true, []string{"info|name|extension.yaml|the extension is installed as 'myanalytics', but declares the name 'analytics'"}},
		{"installed under the declared name", "c8y-analytics", "name: analytics\n", true, nil},
		{"name of an extension command", "c8y-analytics", "name: analytics\n", false, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := newTestExt(t, tt.folder).file("extension.yaml", tt.manifest, 0644)
			findings := ext.validate(ValidateOptions{SkipFlagCheck: true, RootCommand: testRootCommand(), Installed: tt.installed})
			text := findingsText(findings)
			require.Len(t, findings, len(tt.want), text)
			for _, want := range tt.want {
				assert.Contains(t, text, want)
			}
		})
	}
}

func TestIsBuiltInCommand(t *testing.T) {
	root := testRootCommand()
	assert.True(t, isBuiltInCommand(root, "devices"))
	assert.False(t, isBuiltInCommand(root, "analytics"), "extension commands are not built-in")
	assert.False(t, isBuiltInCommand(root, "unknown"))
}

func TestValidateNameInfoIsNotCounted(t *testing.T) {
	ext := newTestExt(t, "c8y-myanalytics").file("extension.yaml", "name: analytics\n", 0644)
	findings := ext.validate(ValidateOptions{SkipFlagCheck: true, Installed: true})
	require.Len(t, findings, 1)
	errorCount, warningCount := CountFindings(findings)
	assert.Zero(t, errorCount+warningCount, "info findings don't fail --strict")
}
