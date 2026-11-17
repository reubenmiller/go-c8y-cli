package root

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newHelpTestRoot() (*cobra.Command, *cobra.Command) {
	root := &cobra.Command{Use: "c8y"}
	newHelpFlag(root)
	root.PersistentFlags().String("output", "", "Output format")
	root.PersistentFlags().Int("pageSize", 5, "Page size")
	root.PersistentFlags().Bool("customGlobal", false, "Flag without a category")
	annotateHelpCategories(root.PersistentFlags())
	root.SetUsageFunc(newHelpUsageFunc(nil))

	child := &cobra.Command{Use: "list", Run: func(*cobra.Command, []string) {}}
	child.Flags().String("name", "", "Filter by name")
	root.AddCommand(child)
	return root, child
}

func runHelp(t *testing.T, root *cobra.Command, args ...string) string {
	t.Helper()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs(NormalizeHelpArgs(root, args))
	require.NoError(t, root.Execute())
	return out.String()
}

func TestHelpValue(t *testing.T) {
	v := &helpValue{}
	require.NoError(t, v.Set("true"))
	assert.Equal(t, "true", v.String())
	assert.Equal(t, "", v.level)

	require.NoError(t, v.Set("local,output"))
	assert.Equal(t, "true", v.String())
	assert.Equal(t, "local,output", v.level)

	require.NoError(t, v.Set("false"))
	assert.Equal(t, "false", v.String())

	assert.Error(t, v.Set("unknown"))
	assert.Equal(t, "bool", v.Type())
}

func TestIsHelpLevel(t *testing.T) {
	for _, value := range []string{"all", "local", "global", "category", "categories", "output", "local,output", "LOCAL, Output"} {
		assert.True(t, isHelpLevel(value), value)
	}
	for _, value := range []string{"", "devices", "local,devices", "12345"} {
		assert.False(t, isHelpLevel(value), value)
	}
}

func TestParseHelpLevel(t *testing.T) {
	sel := parseHelpLevel("")
	assert.True(t, sel.showsAllFlags())

	sel = parseHelpLevel("invalid")
	assert.True(t, sel.showsAllFlags())

	sel = parseHelpLevel("local")
	assert.True(t, sel.local)
	assert.False(t, sel.showsCategory("output"))

	sel = parseHelpLevel("local,output")
	assert.True(t, sel.local)
	assert.True(t, sel.showsCategory("output"))
	assert.False(t, sel.showsCategory("pagination"))

	sel = parseHelpLevel("global")
	assert.False(t, sel.local)
	assert.True(t, sel.showsCategory("pagination"))
}

func TestNormalizeHelpArgs(t *testing.T) {
	root, _ := newHelpTestRoot()
	cases := map[string]string{
		"list --help local":       "list --help=local",
		"list -h local,output":    "list --help=local,output",
		"list --help":             "list --help",
		"list --help --name x":    "list --help --name x",
		"list --help devices":     "list --help devices",
		"list -- --help local":    "list -- --help local",
		"list --name x -h global": "list --name x --help=global",
	}
	for args, want := range cases {
		assert.Equal(t, strings.Fields(want), NormalizeHelpArgs(root, strings.Fields(args)), args)
	}

	// extension commands handle their own help
	ext := &cobra.Command{Use: "ext", DisableFlagParsing: true, Run: func(*cobra.Command, []string) {}}
	root.AddCommand(ext)
	assert.Equal(t, []string{"ext", "--help", "local"}, NormalizeHelpArgs(root, []string{"ext", "--help", "local"}))
}

func TestHelpLevels(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		root, _ := newHelpTestRoot()
		out := runHelp(t, root, "list", "--help")
		assert.Contains(t, out, "--name")
		assert.Contains(t, out, "Global Flags (output):")
		assert.Contains(t, out, "Global Flags (pagination):")
		assert.Contains(t, out, "Global Flags (other):")
		assert.NotContains(t, out, "--help all\" to show all flags")
	})

	t.Run("local", func(t *testing.T) {
		root, _ := newHelpTestRoot()
		out := runHelp(t, root, "list", "--help", "local")
		assert.Contains(t, out, "--name")
		assert.NotContains(t, out, "Global Flags")
		assert.Contains(t, out, "c8y list --help all")
	})

	t.Run("global", func(t *testing.T) {
		root, _ := newHelpTestRoot()
		out := runHelp(t, root, "list", "--help=global")
		assert.NotContains(t, out, "--name")
		assert.Contains(t, out, "Global Flags (output):")
	})

	t.Run("category filter", func(t *testing.T) {
		root, _ := newHelpTestRoot()
		out := runHelp(t, root, "list", "-h", "local,pagination")
		assert.Contains(t, out, "--name")
		assert.Contains(t, out, "--pageSize")
		assert.NotContains(t, out, "--output")
	})

	t.Run("list categories", func(t *testing.T) {
		root, _ := newHelpTestRoot()
		out := runHelp(t, root, "list", "--help", "category")
		assert.Contains(t, out, "Global Flag Categories:")
		assert.NotContains(t, out, "--name")
	})

	t.Run("root command", func(t *testing.T) {
		root, _ := newHelpTestRoot()
		out := runHelp(t, root, "--help", "global")
		assert.Contains(t, out, "Global Flags (output):")
		assert.Contains(t, out, "Available Commands:")
	})
}
