package root

import (
	"sort"
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
		Annotations:        map[string]string{annotationOutput: "json"},
		RunE:               record,
	})
	services.AddCommand(&cobra.Command{
		Use:                "raw",
		DisableFlagParsing: true,
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

func TestExtractOutputFlags(t *testing.T) {
	cases := []struct {
		args      string
		wantFlags string
		wantRest  string
	}{
		{"--name x -o csv --select id", "-o csv --select id", "--name x"},
		{"--output=table --compact --filter name=a* --name x", "--output=table --compact --filter name=a*", "--name x"},
		{"-ocsv -c -r -M --view off", "-ocsv -c -r -M --view off", ""},
		{"--name x -- --output csv", "", "--name x -- --output csv"},
		{"--outputFile x --selected y", "", "--outputFile x --selected y"},
		{"-v --verbose", "", "-v --verbose"},
	}
	for _, tc := range cases {
		flags, rest := extractOutputFlags(strings.Fields(tc.args))
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
			name:       "output flags of json commands are parsed and removed",
			args:       "-v foo services list --name x -o csv --select id,name",
			want:       "foo services list --name x",
			wantFlags:  map[string]string{"verbose": "true", "output": "csv"},
			wantSelect: []string{"id,name"},
		},
		{
			name:      "output flags of raw commands are passed on",
			args:      "foo services raw -o csv",
			want:      "foo services raw -o csv",
			wantFlags: map[string]string{"output": "json"},
		},
		{
			name:      "output flags of mounts are passed on",
			args:      "foo report -o csv",
			want:      "foo report -o csv",
			wantFlags: map[string]string{"output": "json"},
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

func TestMergeEnv(t *testing.T) {
	base := []string{"PATH=/bin", "C8Y_HOST=https://old", "C8Y_TOKEN=old", "C8Y_PASSWORD=keep", "OTHER=1"}
	env := mergeEnv(base,
		map[string]string{"C8Y_HOST": "https://new", "C8Y_EXTENSION_NAME": "foo"},
		map[string]bool{"C8Y_HOST": true, "C8Y_TOKEN": true},
	)
	sort.Strings(env)
	assert.Equal(t, []string{
		"C8Y_EXTENSION_NAME=foo",
		"C8Y_HOST=https://new",
		"C8Y_PASSWORD=keep",
		"OTHER=1",
		"PATH=/bin",
	}, env)
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
