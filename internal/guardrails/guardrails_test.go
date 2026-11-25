// Package guardrails holds tests that keep every go-c8y-cli command on the
// same shape: cross-cutting behaviour (confirmation, dry run, pagination,
// reference resolution, rendering) lives in the c8ystream Runner and the
// go-c8y SDK, never in a single command.
//
// The checks ratchet: existing violations are listed in allowlist.json and
// only new ones fail. A listed violation that has been fixed must be removed
// from the list too, so the list only ever shrinks. Regenerate it with
//
//	UPDATE_GUARDRAILS=1 go test ./internal/guardrails/
//
// and review the diff: an entry should only ever disappear.
package guardrails

import (
	"encoding/json"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/clibuild"
	"github.com/reubenmiller/go-c8y-cli/v2/internal/clisurface"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

const allowlistFile = "allowlist.json"

// sourceRules are patterns no command under pkg/cmd may use. Each names the
// central replacement.
var sourceRules = []struct {
	Name    string
	Pattern *regexp.Regexp
	Why     string
}{
	{"submit-wrapper", regexp.MustCompile(`c8ystream\.Submit(Status|Upload|Value)?\(`),
		"return the SDK result from the Call; the Runner's request approver confirms every request"},
	{"result-adapter", regexp.MustCompile(`c8ystream\.(FromResult|FromStatus|FromJSONValue|ListCall)\(`),
		"return the SDK result (op.Result or iterator) from the Call; the Runner renders it"},
	{"pagination-literal", regexp.MustCompile(`pagination\.PaginationOptions\{`),
		"use r.Pagination()"},
	{"dry-run-check", regexp.MustCompile(`IsDryRun\(`),
		"dry-run results render nothing by themselves (op.Result.Documents)"},
	{"cli-body-ref", regexp.MustCompile(`\.Resolve(SourceID|BodyRef)\(`),
		"the SDK resolves references in request bodies (ResolveBodyRefs)"},
	{"go-c8y-v1", regexp.MustCompile(`"github\.com/reubenmiller/go-c8y/pkg/c8y"`),
		"use the go-c8y v2 SDK"},
	{"v1-worker", regexp.MustCompile(`RunWith(Generic)?Workers\(|c8yfetcher\.`),
		"use the c8ystream Runner"},
}

// sharedFlagNames are flags that mean the same thing on every command, so they
// must have the same type everywhere.
var sharedFlagNames = []string{"device", "id", "group", "agent"}

// singleValuedFlags are the commands where a shared flag name legitimately
// takes a different type, because there it does not reference existing objects
// (which is what the list type is for). Each entry says why. An entry which no
// longer differs from the convention fails the test, so the list only shrinks.
var singleValuedFlags = map[string]map[string]string{
	"id": {
		"tenants create": "the id of the tenant to create, not a reference",
		"devices enroll": "the external id of the device to enroll, not a reference",
	},
	"group": {
		"deviceregistration register-basic":       "a group path (main_group/sub_group), created if it does not exist; a list would split names with commas",
		"deviceregistration register-ca":          "a group path (main_group/sub_group), created if it does not exist; a list would split names with commas",
		"deviceregistration register-external-ca": "a group path (main_group/sub_group), created if it does not exist; a list would split names with commas",
	},
}

// localCommandGroups are the top-level command groups which only manage local
// state, so their commands may be marked with cmdutil.MarkLocalCommand (and
// are then exempt from the session mode check).
var localCommandGroups = map[string]bool{
	"alias":      true,
	"cache":      true,
	"extensions": true,
	"sessions":   true,
	"settings":   true,
}

type allowlist map[string][]string

func readAllowlist(t *testing.T) allowlist {
	t.Helper()
	b, err := os.ReadFile(allowlistFile)
	if os.IsNotExist(err) {
		return allowlist{}
	}
	if err != nil {
		t.Fatal(err)
	}
	var a allowlist
	if err := json.Unmarshal(b, &a); err != nil {
		t.Fatal(err)
	}
	return a
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// ratchet compares the found violations of each rule with the allowlist.
func ratchet(t *testing.T, found map[string][]string, hint func(rule string) string) {
	t.Helper()
	allowed := readAllowlist(t)
	if os.Getenv("UPDATE_GUARDRAILS") != "" {
		merged := allowlist{}
		for k, v := range allowed {
			merged[k] = v
		}
		for rule := range found {
			merged[rule] = found[rule]
		}
		for rule, v := range merged {
			if len(v) == 0 {
				delete(merged, rule)
			}
		}
		writeAllowlist(t, merged)
		return
	}
	for rule, items := range found {
		for _, item := range items {
			if !slices.Contains(allowed[rule], item) {
				t.Errorf("%s: %s (%s)", rule, item, hint(rule))
			}
		}
	}
	for rule := range found {
		for _, item := range allowed[rule] {
			if !slices.Contains(found[rule], item) {
				t.Errorf("%s: %s is fixed; remove it from %s", rule, item, allowlistFile)
			}
		}
	}
}

func writeAllowlist(t *testing.T, a allowlist) {
	t.Helper()
	for k := range a {
		sort.Strings(a[k])
	}
	b, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(allowlistFile, append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCommandSource(t *testing.T) {
	root := repoRoot(t)
	found := map[string][]string{}
	for _, r := range sourceRules {
		found[r.Name] = []string{}
	}
	err := filepath.WalkDir(filepath.Join(root, "pkg", "cmd"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		for _, r := range sourceRules {
			if r.Pattern.Match(b) {
				found[r.Name] = append(found[r.Name], rel)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	ratchet(t, found, func(rule string) string {
		for _, r := range sourceRules {
			if r.Name == rule {
				return r.Why
			}
		}
		return ""
	})
}

func commandTree(t *testing.T) *cobra.Command {
	t.Helper()
	// Build only the built-in commands: no installed extensions (user or
	// system wide) and no user settings, so the result is the same on every
	// machine.
	empty := t.TempDir()
	t.Setenv("C8Y_SESSION_HOME", empty)
	t.Setenv("C8Y_SETTINGS_EXTENSIONS_DATADIR", empty)
	t.Setenv("C8Y_SETTINGS_EXTENSIONS_SYSTEMDIRS", empty)
	root, err := clibuild.NewRootCommand()
	if err != nil {
		t.Fatal(err)
	}
	return root.Command
}

func walk(cmd *cobra.Command, fn func(*cobra.Command)) {
	for _, c := range cmd.Commands() {
		if c.Runnable() {
			fn(c)
		}
		walk(c, fn)
	}
}

func path(cmd *cobra.Command) string {
	return strings.TrimPrefix(cmd.CommandPath(), cmd.Root().Name()+" ")
}

// TestMutatingCommandsAreModeGuarded checks that every command that changes
// data checks the session mode (create/update/delete mode) before running.
// Commands which only change local state (aliases, cache, extensions,
// sessions, settings) are marked with cmdutil.MarkLocalCommand instead: the
// mode protects tenant data, and e.g. switching sessions must work from a
// read-only session.
func TestMutatingCommandsAreModeGuarded(t *testing.T) {
	found := map[string][]string{"mode-guard": {}}
	walk(commandTree(t), func(c *cobra.Command) {
		if cmdutil.IsLocalCommand(c) {
			// The annotation skips this check, so it must not become a way
			// around the ratchet: only the command groups which manage local
			// state may use it.
			if group, _, _ := strings.Cut(path(c), " "); !localCommandGroups[group] {
				t.Errorf("%s: only commands in %v may be marked with cmdutil.MarkLocalCommand; check the session mode instead", path(c), slices.Sorted(maps.Keys(localCommandGroups)))
			}
			return
		}
		method := c.Annotations["semanticMethod"]
		if method == "" {
			method = clisurface.DeriveMethod(c.Name())
		}
		if method == "GET" {
			return
		}
		if c.PreRunE == nil && c.PreRun == nil {
			found["mode-guard"] = append(found["mode-guard"], path(c))
		}
	})
	ratchet(t, found, func(string) string {
		return "add PreRunE: f.CreateModeEnabled/UpdateModeEnabled/DeleteModeEnabled (or cmdutil.MarkLocalCommand if it only changes local state)"
	})
}

// TestSharedFlagTypes checks that flags which mean the same thing on every
// command (device, id, group, agent) are declared with the same type, apart
// from the documented exceptions in singleValuedFlags.
func TestSharedFlagTypes(t *testing.T) {
	types := map[string]map[string][]string{} // flag -> type -> commands
	walk(commandTree(t), func(c *cobra.Command) {
		for _, name := range sharedFlagNames {
			f := c.Flags().Lookup(name)
			if f == nil {
				continue
			}
			if types[name] == nil {
				types[name] = map[string][]string{}
			}
			types[name][f.Value.Type()] = append(types[name][f.Value.Type()], path(c))
		}
	})
	found := map[string][]string{}
	for _, name := range sharedFlagNames {
		rule := "flag-type-" + name
		found[rule] = []string{}
		// the most common type is the convention
		major, n := "", -1
		for typ, cmds := range types[name] {
			if len(cmds) > n {
				major, n = typ, len(cmds)
			}
		}
		exceptions := map[string]bool{}
		for typ, cmds := range types[name] {
			if typ == major {
				for _, c := range cmds {
					if _, ok := singleValuedFlags[name][c]; ok {
						t.Errorf("%s: --%s is %s like the other commands now; remove it from singleValuedFlags", c, name, typ)
					}
				}
				continue
			}
			for _, c := range cmds {
				if _, ok := singleValuedFlags[name][c]; ok {
					exceptions[c] = true
					continue
				}
				found[rule] = append(found[rule], c+" ("+typ+", expected "+major+")")
			}
		}
		for c := range singleValuedFlags[name] {
			if !exceptions[c] && !slices.Contains(types[name][major], c) {
				t.Errorf("%s: has no --%s flag any more; remove it from singleValuedFlags", c, name)
			}
		}
	}
	ratchet(t, found, func(string) string { return "use the same flag type as the other commands" })
}

// TestNoGoC8yV1Module keeps the go-c8y v1 SDK out of the CLI: no module of the repository may
// require it, and no Go file may import it. Only the v2 module (github.com/reubenmiller/go-c8y/v2)
// is allowed.
func TestNoGoC8yV1Module(t *testing.T) {
	root := repoRoot(t)
	requiresV1 := regexp.MustCompile(`(?m)^\s*(require\s+)?github\.com/reubenmiller/go-c8y\s+v`)
	importsV1 := regexp.MustCompile(`"github\.com/reubenmiller/go-c8y/pkg/`)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", ".claude":
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		switch {
		case d.Name() == "go.mod":
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if requiresV1.Match(b) {
				t.Errorf("%s requires the go-c8y v1 module; use github.com/reubenmiller/go-c8y/v2", rel)
			}
		case strings.HasSuffix(d.Name(), ".go") && rel != filepath.Join("internal", "guardrails", "guardrails_test.go"):
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if importsV1.Match(b) {
				t.Errorf("%s imports the go-c8y v1 SDK; use the v2 SDK", rel)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
