package cmdtree

import (
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// leaf creates a test leaf. When run, the command records its origin in *ran
func leaf(source SourceKind, origin, name string, ran *string) Leaf {
	return Leaf{
		Source: source,
		Origin: origin,
		Build: func() (*cobra.Command, error) {
			return &cobra.Command{
				Use: name,
				RunE: func(cmd *cobra.Command, args []string) error {
					if ran != nil {
						*ran = origin
					}
					return nil
				},
			}, nil
		},
	}
}

func split(path string) []string {
	return strings.Split(path, " ")
}

func addLeaf(t *Tree, source SourceKind, origin, path string, ran *string) {
	parts := split(path)
	t.AddLeaf(parts, leaf(source, origin, parts[len(parts)-1], ran))
}

// recorder records which command ran and any conflicts
type recorder struct {
	ran       string
	conflicts []Conflict
}

func newTree(r *recorder) *Tree {
	tree := New("analytics", "analytics extension")
	tree.OnConflict = func(c Conflict) {
		r.conflicts = append(r.conflicts, c)
	}
	return tree
}

func execute(t *testing.T, tree *Tree, r *recorder, args string) {
	t.Helper()
	cmd, err := tree.ToCobra()
	if err != nil {
		t.Fatalf("unexpected error building tree: %s", err)
	}
	if cmd == nil {
		t.Fatalf("tree is empty")
	}
	root := &cobra.Command{Use: "c8y", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(cmd)
	root.SetArgs(split(args))
	r.ran = ""
	if err := root.Execute(); err != nil {
		t.Fatalf("command failed. args=%s, err=%s", args, err)
	}
}

func childNames(cmd *cobra.Command) []string {
	names := []string{}
	for _, c := range cmd.Commands() {
		names = append(names, c.Name())
	}
	sort.Strings(names)
	return names
}

func findCmd(t *testing.T, tree *Tree, path string) *cobra.Command {
	t.Helper()
	cmd, err := tree.ToCobra()
	if err != nil {
		t.Fatalf("unexpected error building tree: %s", err)
	}
	found, _, err := cmd.Find(split(path))
	if err != nil {
		t.Fatalf("command not found: %s", path)
	}
	return found
}

func TestMergeAPIGroupAndScriptFolder(t *testing.T) {
	// Issue #585
	r := &recorder{}
	tree := newTree(r)

	tree.AddGroup(split("management"), "Manage analytics", "", "api/management.yaml")
	addLeaf(tree, SourceAPI, "api/management.yaml", "management health", &r.ran)
	addLeaf(tree, SourceScript, "commands/management/logs", "management logs", &r.ran)

	execute(t, tree, r, "analytics management health")
	if r.ran != "api/management.yaml" {
		t.Errorf("expected api command to run. got=%q", r.ran)
	}

	execute(t, tree, r, "analytics management logs")
	if r.ran != "commands/management/logs" {
		t.Errorf("expected script command to run. got=%q", r.ran)
	}

	cmd, _ := tree.ToCobra()
	if got := childNames(cmd); strings.Join(got, ",") != "management" {
		t.Errorf("expected exactly one management group. got=%v", got)
	}
	if got := childNames(findCmd(t, tree, "management")); strings.Join(got, ",") != "health,logs" {
		t.Errorf("unexpected management children. got=%v", got)
	}
	if len(r.conflicts) != 0 {
		t.Errorf("expected no conflicts. got=%v", r.conflicts)
	}
}

func TestMergeScriptBeforeAPI(t *testing.T) {
	// Order of registration must not matter
	r := &recorder{}
	tree := newTree(r)

	addLeaf(tree, SourceScript, "commands/management/logs", "management logs", &r.ran)
	tree.AddGroup(split("management"), "Manage analytics", "", "api/management.yaml")
	addLeaf(tree, SourceAPI, "api/management.yaml", "management health", &r.ran)

	if got := childNames(findCmd(t, tree, "management")); strings.Join(got, ",") != "health,logs" {
		t.Errorf("unexpected management children. got=%v", got)
	}
	if got := findCmd(t, tree, "management").Short; got != "Manage analytics" {
		t.Errorf("expected explicit group description. got=%q", got)
	}
}

func TestDuplicateSpecGroups(t *testing.T) {
	r := &recorder{}
	tree := newTree(r)

	tree.AddGroup(split("devices"), "Devices from a", "", "api/a.yaml")
	addLeaf(tree, SourceAPI, "api/a.yaml", "devices list", &r.ran)
	tree.AddGroup(split("devices"), "Devices from b", "", "api/b.yaml")
	addLeaf(tree, SourceAPI, "api/b.yaml", "devices get", &r.ran)

	execute(t, tree, r, "analytics devices list")
	if r.ran != "api/a.yaml" {
		t.Errorf("unexpected command ran. got=%q", r.ran)
	}
	execute(t, tree, r, "analytics devices get")
	if r.ran != "api/b.yaml" {
		t.Errorf("unexpected command ran. got=%q", r.ran)
	}

	cmd, _ := tree.ToCobra()
	if got := childNames(cmd); strings.Join(got, ",") != "devices" {
		t.Errorf("expected exactly one devices group. got=%v", got)
	}
	if got := findCmd(t, tree, "devices").Short; got != "Devices from a" {
		t.Errorf("expected first explicit description to win. got=%q", got)
	}
}

func TestDeepNesting(t *testing.T) {
	r := &recorder{}
	tree := newTree(r)

	addLeaf(tree, SourceScript, "commands/hello", "hello", &r.ran)
	addLeaf(tree, SourceScript, "commands/a/b/c/run", "a b c run", &r.ran)

	execute(t, tree, r, "analytics hello")
	if r.ran != "commands/hello" {
		t.Errorf("unexpected command ran. got=%q", r.ran)
	}
	execute(t, tree, r, "analytics a b c run")
	if r.ran != "commands/a/b/c/run" {
		t.Errorf("unexpected command ran. got=%q", r.ran)
	}

	for path, want := range map[string]string{"a": "b", "a b": "c", "a b c": "run"} {
		if got := childNames(findCmd(t, tree, path)); strings.Join(got, ",") != want {
			t.Errorf("unexpected children. path=%s, got=%v, want=%s", path, got, want)
		}
	}
	if got := findCmd(t, tree, "a b").Short; got != "b command group" {
		t.Errorf("expected generated description. got=%q", got)
	}
}

func TestNestedAPIGroups(t *testing.T) {
	// Issue #562: api specs can define nested groups, and a spec without commands
	// can provide the description of a parent group
	r := &recorder{}
	tree := newTree(r)

	tree.AddGroup(split("hierarchy items"), "Manage items", "", "api/hierarchy_items.yaml")
	addLeaf(tree, SourceAPI, "api/hierarchy_items.yaml", "hierarchy items list", &r.ran)
	addLeaf(tree, SourceAPI, "api/hierarchy_items.yaml", "hierarchy items get", &r.ran)
	tree.AddGroup(split("hierarchy"), "Nested groups", "", "api/hierarchy.yaml")

	execute(t, tree, r, "analytics hierarchy items get")
	if r.ran != "api/hierarchy_items.yaml" {
		t.Errorf("unexpected command ran. got=%q", r.ran)
	}
	if got := findCmd(t, tree, "hierarchy").Short; got != "Nested groups" {
		t.Errorf("expected parent group description from separate spec. got=%q", got)
	}
	if got := findCmd(t, tree, "hierarchy items").Short; got != "Manage items" {
		t.Errorf("unexpected nested group description. got=%q", got)
	}
	if got := childNames(findCmd(t, tree, "hierarchy items")); strings.Join(got, ",") != "get,list" {
		t.Errorf("unexpected children. got=%v", got)
	}
}

func TestLeafConflictPrecedence(t *testing.T) {
	for _, apiFirst := range []bool{true, false} {
		r := &recorder{}
		tree := newTree(r)

		addAPI := func() { addLeaf(tree, SourceAPI, "api/health.yaml", "management health", &r.ran) }
		addScript := func() {
			addLeaf(tree, SourceScript, "commands/management/health", "management health", &r.ran)
		}
		if apiFirst {
			addAPI()
			addScript()
		} else {
			addScript()
			addAPI()
		}

		execute(t, tree, r, "analytics management health")
		if r.ran != "api/health.yaml" {
			t.Errorf("expected api command to win. apiFirst=%v, got=%q", apiFirst, r.ran)
		}
		if len(r.conflicts) != 1 {
			t.Fatalf("expected one conflict. apiFirst=%v, got=%v", apiFirst, r.conflicts)
		}
		want := Conflict{
			Path:    "analytics management health",
			Kept:    "api/health.yaml",
			Dropped: "commands/management/health",
		}
		if r.conflicts[0] != want {
			t.Errorf("unexpected conflict. got=%v, want=%v", r.conflicts[0], want)
		}
	}
}

func TestLeafConflictSameSourceKeepsFirst(t *testing.T) {
	r := &recorder{}
	tree := newTree(r)
	addLeaf(tree, SourceAPI, "api/a.yaml", "devices list", &r.ran)
	addLeaf(tree, SourceAPI, "api/b.yaml", "devices list", &r.ran)

	execute(t, tree, r, "analytics devices list")
	if r.ran != "api/a.yaml" {
		t.Errorf("expected first command to win. got=%q", r.ran)
	}
	if len(r.conflicts) != 1 || r.conflicts[0].Dropped != "api/b.yaml" {
		t.Errorf("unexpected conflicts. got=%v", r.conflicts)
	}
}

func TestGroupOverLeaf(t *testing.T) {
	for _, groupFirst := range []bool{true, false} {
		r := &recorder{}
		tree := newTree(r)

		addGroup := func() {
			tree.AddGroup(split("management"), "Manage", "", "api/management.yaml")
			addLeaf(tree, SourceAPI, "api/management.yaml", "management health", &r.ran)
		}
		addScript := func() {
			addLeaf(tree, SourceScript, "commands/management", "management", &r.ran)
		}
		if groupFirst {
			addGroup()
			addScript()
		} else {
			addScript()
			addGroup()
		}

		mgmt := findCmd(t, tree, "management")
		if mgmt.Runnable() {
			t.Errorf("expected management to be a group. groupFirst=%v", groupFirst)
		}
		execute(t, tree, r, "analytics management health")
		if r.ran != "api/management.yaml" {
			t.Errorf("unexpected command ran. groupFirst=%v, got=%q", groupFirst, r.ran)
		}
		if len(r.conflicts) != 1 {
			t.Fatalf("expected one conflict. groupFirst=%v, got=%v", groupFirst, r.conflicts)
		}
		if r.conflicts[0].Dropped != "commands/management" || r.conflicts[0].Path != "analytics management" {
			t.Errorf("unexpected conflict. groupFirst=%v, got=%v", groupFirst, r.conflicts[0])
		}
	}
}

func TestEmptyTree(t *testing.T) {
	tree := New("viewsonly", "viewsonly extension")
	tree.AddGroup(split("empty"), "Empty group", "", "api/empty.yaml")

	if !tree.IsEmpty() {
		t.Errorf("expected tree to be empty")
	}
	cmd, err := tree.ToCobra()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if cmd != nil {
		t.Errorf("expected no command for an empty tree")
	}
}

func TestEmptyGroupsAreOmitted(t *testing.T) {
	r := &recorder{}
	tree := newTree(r)
	tree.AddGroup(split("empty"), "Empty group", "", "api/empty.yaml")
	addLeaf(tree, SourceScript, "commands/hello", "hello", &r.ran)

	cmd, _ := tree.ToCobra()
	if got := childNames(cmd); strings.Join(got, ",") != "hello" {
		t.Errorf("expected empty group to be omitted. got=%v", got)
	}
}

func TestBuildErrorsAreCollected(t *testing.T) {
	r := &recorder{}
	tree := newTree(r)
	addLeaf(tree, SourceScript, "commands/good", "good", &r.ran)
	tree.AddLeaf(split("bad"), Leaf{
		Source: SourceAPI,
		Origin: "api/bad.yaml",
		Build: func() (*cobra.Command, error) {
			return nil, errors.New("boom")
		},
	})

	cmd, err := tree.ToCobra()
	if err == nil || !strings.Contains(err.Error(), "api/bad.yaml") {
		t.Errorf("expected error mentioning the origin. got=%v", err)
	}
	if got := childNames(cmd); strings.Join(got, ",") != "good" {
		t.Errorf("expected remaining commands to be built. got=%v", got)
	}
}

func TestDeterministicOutput(t *testing.T) {
	build := func(order []int) (string, []Conflict) {
		r := &recorder{}
		tree := newTree(r)
		adds := []func(){
			func() { tree.AddGroup(split("management"), "Manage", "", "api/m.yaml") },
			func() { addLeaf(tree, SourceAPI, "api/m.yaml", "management health", nil) },
			func() { addLeaf(tree, SourceScript, "commands/management/health", "management health", nil) },
			func() { addLeaf(tree, SourceScript, "commands/management/logs", "management logs", nil) },
			func() { addLeaf(tree, SourceScript, "commands/z/y", "z y", nil) },
		}
		for _, i := range order {
			adds[i]()
		}
		cmd, _ := tree.ToCobra()
		return dump(cmd, ""), r.conflicts
	}

	want, wantConflicts := build([]int{0, 1, 2, 3, 4})
	got, gotConflicts := build([]int{4, 3, 2, 1, 0})
	if got != want {
		t.Errorf("command tree depends on registration order.\nwant:\n%s\ngot:\n%s", want, got)
	}
	if len(gotConflicts) != 1 || len(wantConflicts) != 1 || gotConflicts[0] != wantConflicts[0] {
		t.Errorf("conflicts depend on registration order. want=%v, got=%v", wantConflicts, gotConflicts)
	}
}

func dump(cmd *cobra.Command, indent string) string {
	var b strings.Builder
	writeTree(&b, cmd, indent)
	return b.String()
}

func writeTree(b *strings.Builder, cmd *cobra.Command, indent string) {
	b.WriteString(indent + cmd.Name() + " (" + cmd.Short + ")\n")
	for _, c := range cmd.Commands() {
		writeTree(b, c, indent+"  ")
	}
}
