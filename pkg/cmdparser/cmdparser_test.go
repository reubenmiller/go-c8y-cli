package cmdparser

import (
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var update = flag.Bool("update", false, "update golden files")

const kitchensinkAPIDir = "../../tests/testdata/extensions/c8y-kitchensink/api"

func newTestFactory() *cmdutil.Factory {
	cfg := config.NewConfig(viper.New())
	log := logger.NewDummyLogger("test")
	ios, _, _, _ := iostreams.Test()
	return &cmdutil.Factory{
		IOStreams: ios,
		Config:    func() (*config.Config, error) { return cfg, nil },
		Logger:    func() (*logger.Logger, error) { return log, nil },
	}
}

func describeCommand(b *strings.Builder, cmd *cobra.Command, indent string) {
	b.WriteString(indent + "cmd: " + cmd.Name() + "\n")
	b.WriteString(indent + "  short: " + cmd.Short + "\n")
	b.WriteString(indent + "  hidden: " + boolString(cmd.Hidden) + "\n")
	b.WriteString(indent + "  runnable: " + boolString(cmd.Runnable()) + "\n")

	flags := []string{}
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		flags = append(flags, "--"+f.Name+" ("+f.Value.Type()+") default="+f.DefValue)
	})
	sort.Strings(flags)
	for _, f := range flags {
		b.WriteString(indent + "  flag: " + f + "\n")
	}

	children := cmd.Commands()
	sort.Slice(children, func(i, j int) bool { return children[i].Name() < children[j].Name() })
	for _, c := range children {
		describeCommand(b, c, indent+"  ")
	}
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

// TestParseCommandKitchensink checks that the commands created from the kitchensink
// extension's API specifications do not change unexpectedly.
// Run with -update to regenerate the golden file.
func TestParseCommandKitchensink(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(kitchensinkAPIDir, "*.yaml"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no spec files found. err=%v", err)
	}
	sort.Strings(files)

	var b strings.Builder
	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		cmd, err := ParseCommand(f, newTestFactory(), &cobra.Command{Use: "c8y"})
		f.Close()
		if err != nil {
			t.Fatalf("failed to parse %s: %s", path, err)
		}
		b.WriteString("file: " + filepath.Base(path) + "\n")
		if cmd == nil {
			b.WriteString("  skipped\n")
			continue
		}
		describeCommand(&b, cmd, "  ")
	}

	golden := filepath.Join("testdata", "kitchensink.golden")
	if *update {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, []byte(b.String()), 0644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("missing golden file, run with -update. %s", err)
	}
	if got := b.String(); got != string(want) {
		t.Errorf("parsed commands differ from %s (run with -update if the change is intended)\n%s", golden, diffLines(string(want), got))
	}
}

// TestParseSpecMatchesParseCommand checks that building commands individually with
// ParseSpec and NewCommandFromSpec gives the same commands as ParseCommand
func TestParseSpecMatchesParseCommand(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join(kitchensinkAPIDir, "*.yaml"))
	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		group, err := ParseCommand(f, newTestFactory(), &cobra.Command{Use: "c8y"})
		f.Close()
		if err != nil {
			t.Fatal(err)
		}

		f, _ = os.Open(path)
		spec, err := ParseSpec(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}

		individual := &cobra.Command{Use: spec.Group.Name, Short: spec.Group.Description}
		for _, item := range spec.Commands {
			if item.ShouldIgnore() {
				continue
			}
			cmd, err := NewCommandFromSpec(item, newTestFactory(), path)
			if err != nil {
				t.Fatal(err)
			}
			individual.AddCommand(cmd)
		}

		var want, got strings.Builder
		describeCommand(&want, group, "")
		describeCommand(&got, individual, "")
		if want.String() != got.String() {
			t.Errorf("commands differ for %s\n%s", path, diffLines(want.String(), got.String()))
		}
	}
}

func diffLines(want, got string) string {
	w := strings.Split(want, "\n")
	g := strings.Split(got, "\n")
	var b strings.Builder
	for i := 0; i < len(w) || i < len(g); i++ {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			b.WriteString("line " + strconv.Itoa(i+1) + ":\n  want: " + wl + "\n  got:  " + gl + "\n")
			if b.Len() > 2000 {
				break
			}
		}
	}
	return b.String()
}
