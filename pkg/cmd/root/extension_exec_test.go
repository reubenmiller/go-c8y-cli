package root

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRoot creates a root command with some of the c8y global flags, and an extension
// with a raw command, a json command and a mount
func newTestRoot() (*cobra.Command, *[]string) {
	got := &[]string{}
	record := func(cmd *cobra.Command, args []string) error {
		*got = append([]string{cmd.CommandPath()}, args...)
		return nil
	}
	root := &cobra.Command{Use: "c8y", SilenceErrors: true, SilenceUsage: true}
	root.PersistentFlags().String("session", "", "")
	root.PersistentFlags().Bool("dry", false, "")
	root.PersistentFlags().BoolP("verbose", "v", false, "")
	root.PersistentFlags().StringP("output", "o", "json", "")
	root.PersistentFlags().StringArray("select", nil, "")
	root.PersistentFlags().StringArray("filter", nil, "")
	root.PersistentFlags().String("view", "", "")
	root.PersistentFlags().BoolP("compact", "c", false, "")
	root.PersistentFlags().BoolP("raw", "r", false, "")
	root.PersistentFlags().BoolP("noColor", "M", false, "")
	root.PersistentFlags().IntP("pageSize", "p", 5, "")
	root.PersistentFlags().Bool("examples", false, "")

	root.AddCommand(&cobra.Command{Use: "devices", RunE: record})

	ext := &cobra.Command{
		Use:                "foo",
		Annotations:        map[string]string{annotationExtension: "foo"},
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		RunE:               record,
	}
	services := &cobra.Command{Use: "services"}
	services.AddCommand(&cobra.Command{
		Use:                "list",
		DisableFlagParsing: true,
		Annotations:        map[string]string{annotationOutput: "json", annotationDeclared: "true"},
		RunE:               record,
	})
	services.AddCommand(&cobra.Command{
		Use:                "raw",
		DisableFlagParsing: true,
		Annotations:        map[string]string{annotationDeclared: "true"},
		RunE:               record,
	})
	ext.AddCommand(services)
	root.AddCommand(ext)
	return root, got
}

func TestCountLeadingFlags(t *testing.T) {
	root, _ := newTestRoot()
	flags := root.PersistentFlags()
	cases := []struct {
		args string
		want int
	}{
		{"foo list", 0},
		{"--dry foo list", 1},
		{"--session prod.json foo list", 2},
		{"--session=prod.json foo", 1},
		{"-v --session foo foo list", 3},
		{"-o csv foo", 2},
		{"--dry=true foo", 1},
		{"--unknown value foo", 2},
		{"--dry -- foo", 1},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, countLeadingFlags(flags, strings.Fields(tc.args)), tc.args)
	}
}

func TestExtractGlobalFlags(t *testing.T) {
	root, _ := newTestRoot()
	cases := []struct {
		args      string
		wantFlags string
		wantRest  string
	}{
		{"--name x -o csv --select id", "-o csv --select id", "--name x"},
		{"--output=table --compact --filter name=a* --name x", "--output=table --compact --filter name=a*", "--name x"},
		{"-ocsv -c -r -M --view off", "-ocsv -c -r -M --view off", ""},
		{"--session s.json --dry -v --verbose -v=true -p 10", "--session s.json --dry -v --verbose -v=true -p 10", ""},
		{"--name x -- --output csv", "", "--name x -- --output csv"},
		{"--outputFile x --selected y", "", "--outputFile x --selected y"},
		// combined shorthands are not split
		{"-vc", "", "-vc"},
		// help and examples are passed to the binary
		{"--help -h --examples", "", "--help -h --examples"},
	}
	for _, tc := range cases {
		flags, rest := extractGlobalFlags(root.PersistentFlags(), strings.Fields(tc.args))
		assert.Equal(t, tc.wantFlags, strings.Join(flags, " "), tc.args)
		assert.Equal(t, tc.wantRest, strings.Join(rest, " "), tc.args)
	}
}

func TestPrepareExtensionArgs(t *testing.T) {
	cases := []struct {
		name       string
		args       string
		want       string
		wantFlags  map[string]string
		wantSelect []string
	}{
		{
			name: "built-in command is unchanged",
			args: "--dry devices --output csv",
			want: "--dry devices --output csv",
		},
		{
			name:      "leading flags are parsed and removed",
			args:      "--session prod.json --dry foo report --month 10",
			want:      "foo report --month 10",
			wantFlags: map[string]string{"session": "prod.json", "dry": "true"},
		},
		{
			name:       "global flags of json commands are parsed and removed",
			args:       "-v foo services list --name x -o csv --select id,name --session s.json",
			want:       "foo services list --name x",
			wantFlags:  map[string]string{"verbose": "true", "output": "csv", "session": "s.json"},
			wantSelect: []string{"id,name"},
		},
		{
			name:      "global flags of raw commands are parsed and removed",
			args:      "foo services raw --session s.json --dry --x 1 -p 10",
			want:      "foo services raw --x 1",
			wantFlags: map[string]string{"session": "s.json", "dry": "true", "pageSize": "10"},
		},
		{
			name:      "help and examples are passed to declared commands",
			args:      "foo services list --help --examples",
			want:      "foo services list --help --examples",
			wantFlags: map[string]string{"examples": "false"},
		},
		{
			name:      "global flags of mounts are passed on",
			args:      "foo report -o csv --session s.json",
			want:      "foo report -o csv --session s.json",
			wantFlags: map[string]string{"output": "json", "session": ""},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, got := newTestRoot()
			args := PrepareExtensionArgs(root, strings.Fields(tc.args))
			assert.Equal(t, tc.want, strings.Join(args, " "))

			for name, value := range tc.wantFlags {
				assert.Equal(t, value, root.PersistentFlags().Lookup(name).Value.String(), name)
			}
			if tc.wantSelect != nil {
				values, err := root.PersistentFlags().GetStringArray("select")
				require.NoError(t, err)
				assert.Equal(t, tc.wantSelect, values)
			}

			// The prepared args run the expected command
			root.SetArgs(args)
			require.NoError(t, root.Execute())
			require.NotEmpty(t, *got)

			// select is only parsed once
			if tc.wantSelect != nil {
				values, _ := root.PersistentFlags().GetStringArray("select")
				assert.Equal(t, tc.wantSelect, values)
			}
		})
	}
}

func TestPrepareExtensionArgsPassesArgsToCommand(t *testing.T) {
	root, got := newTestRoot()
	root.SetArgs(PrepareExtensionArgs(root, strings.Fields("--dry foo services list --name x --select id")))
	require.NoError(t, root.Execute())
	assert.Equal(t, []string{"c8y foo services list", "--name", "x"}, *got)
}

func TestSplitJSONValues(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		want      []string
		wantArray bool
		wantOK    bool
	}{
		{"empty", "  \n", nil, false, true},
		{"object", `{"id":"1"}`, []string{`{"id":"1"}`}, false, true},
		{"array", `[{"id":"1"}, {"id":"2"}]`, []string{`{"id":"1"}`, `{"id":"2"}`}, true, true},
		{"ndjson", "{\"id\":\"1\"}\n\n{\"id\":\"2\"}\n", []string{`{"id":"1"}`, `{"id":"2"}`}, true, true},
		{"text", "hello", nil, false, false},
		{"partial ndjson", "{\"id\":\"1\"}\nhello\n", nil, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			values, isArray, ok := splitJSONValues([]byte(tc.input))
			assert.Equal(t, tc.wantOK, ok)
			assert.Equal(t, tc.wantArray, isArray)
			got := []string{}
			for _, v := range values {
				got = append(got, string(v))
			}
			if tc.want == nil {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestParseCobraCompletion(t *testing.T) {
	completions, directive := parseCobraCompletion(strings.NewReader("alpha\tfirst\nbeta\n:4\n"))
	assert.Equal(t, []string{"alpha\tfirst", "beta"}, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)

	completions, directive = parseCobraCompletion(strings.NewReader(""))
	assert.Empty(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveDefault, directive)
}

func TestIsExtensionCommand(t *testing.T) {
	root, _ := newTestRoot()
	list, _, err := root.Find([]string{"foo", "services", "list"})
	require.NoError(t, err)
	assert.True(t, IsExtensionCommand(list))

	devices, _, err := root.Find([]string{"devices"})
	require.NoError(t, err)
	assert.False(t, IsExtensionCommand(devices))
}

func TestHasHelpFlag(t *testing.T) {
	cases := map[string]bool{
		"--help":            true,
		"-h":                true,
		"--status x --help": true,
		"--help=true":       true,
		"--status x":        false,
		"-- --help":         false,
		"":                  false,
	}
	for args, want := range cases {
		assert.Equal(t, want, hasHelpFlag(strings.Fields(args)), args)
	}
}

func TestDeclaredCommandCompletion(t *testing.T) {
	root, _ := newTestRoot()
	require.NoError(t, root.RegisterFlagCompletionFunc("output", cobra.FixedCompletions([]string{"json", "csv", "table"}, cobra.ShellCompDirectiveNoFileComp)))
	require.NoError(t, root.RegisterFlagCompletionFunc("session", cobra.FixedCompletions([]string{"dev.json", "prod.json"}, cobra.ShellCompDirectiveNoFileComp)))

	var binaryArgs []string
	list, _, err := root.Find([]string{"foo", "services", "list"})
	require.NoError(t, err)
	list.ValidArgsFunction = declaredCommandCompletion(func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		binaryArgs = append(append([]string{}, args...), toComplete)
		return []cobra.Completion{"--status"}, cobra.ShellCompDirectiveNoFileComp
	})

	complete := func(args ...string) string {
		t.Helper()
		binaryArgs = nil
		var out strings.Builder
		root.SetOut(&out)
		root.SetArgs(append([]string{cobra.ShellCompRequestCmd, "foo", "services", "list"}, args...))
		require.NoError(t, root.Execute())
		return out.String()
	}

	out := complete("--output", "")
	assert.Contains(t, out, "json\ncsv\ntable\n")
	assert.Nil(t, binaryArgs, "binary should not be asked for c8y flag values")

	out = complete("-o", "c")
	assert.Contains(t, out, "csv")

	out = complete("--output=")
	assert.Contains(t, out, "--output=json\n--output=csv\n")

	out = complete("--session", "")
	assert.Contains(t, out, "dev.json\nprod.json\n")
	assert.Nil(t, binaryArgs)

	// Flags without a registered completion don't ask the binary
	complete("--select", "")
	assert.Nil(t, binaryArgs)

	// Other arguments are completed by the binary, without the c8y output flags
	out = complete("--output", "csv", "--session", "dev.json", "--name", "x", "--st")
	assert.Contains(t, out, "--status")
	assert.Equal(t, []string{"--name", "x", "--st"}, binaryArgs)
}
