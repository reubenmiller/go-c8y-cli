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
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/clibuild"
	"github.com/reubenmiller/go-c8y-cli/v2/internal/clisurface"
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
func TestMutatingCommandsAreModeGuarded(t *testing.T) {
	found := map[string][]string{"mode-guard": {}}
	walk(commandTree(t), func(c *cobra.Command) {
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
		return "add PreRunE: f.CreateModeEnabled/UpdateModeEnabled/DeleteModeEnabled"
	})
}

// TestSharedFlagTypes checks that flags which mean the same thing on every
// command (device, id, group, agent) are declared with the same type.
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
		for typ, cmds := range types[name] {
			if typ == major {
				continue
			}
			for _, c := range cmds {
				found[rule] = append(found[rule], c+" ("+typ+", expected "+major+")")
			}
		}
	}
	ratchet(t, found, func(string) string { return "use the same flag type as the other commands" })
}
