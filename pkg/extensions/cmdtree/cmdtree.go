// Package cmdtree builds the command tree of an extension from multiple command sources
// (e.g. API specifications and scripts). Commands are merged by their path, so different
// sources can contribute commands to the same command group.
package cmdtree

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// SourceKind is the kind of source which provides a command.
// Lower values have a higher precedence when two sources define the same command.
type SourceKind int

const (
	// SourceAPI is a command defined in an API specification file (api/*.yaml)
	SourceAPI SourceKind = iota
	// SourceBinary is a command declared in the extension manifest (extension.yaml) which runs the extension's binary
	SourceBinary
	// SourceScript is an executable file under the commands/ folder
	SourceScript
)

func (s SourceKind) String() string {
	switch s {
	case SourceAPI:
		return "api"
	case SourceBinary:
		return "binary"
	case SourceScript:
		return "script"
	}
	return fmt.Sprintf("unknown(%d)", int(s))
}

// Leaf is a runnable command
type Leaf struct {
	// Source of the command, used to resolve conflicts
	Source SourceKind

	// Origin of the command, e.g. a file path. Only used for conflict reporting
	Origin string

	// Build creates the cobra command. The returned command's Use field is set to the
	// last element of the command's path
	Build func() (*cobra.Command, error)
}

// Mount passes all arguments under a command path to a single handler (e.g. the extension's binary).
// A mount is not a leaf: other commands can be added below the mount path, and they take precedence
// over the mount. Any other arguments are passed to the mount.
type Mount struct {
	// Origin of the mount, e.g. a file path. Only used for conflict reporting
	Origin string

	// Short description, used if the mount path is not also a command group with a description
	Short string

	// Long description which is appended to the group's long description
	Long string

	// Configure is called with the cobra command which is created for the mount path. It
	// should set RunE (and optionally ValidArgsFunction). It is called once for the mount
	Configure func(cmd *cobra.Command) error
}

// Conflict describes a command which was dropped because another command
// was registered at the same path
type Conflict struct {
	// Path of the conflicting command (space separated)
	Path string

	// Kept is the origin of the command (or group) which is used
	Kept string

	// Dropped is the origin of the command which is ignored
	Dropped string
}

func (c Conflict) String() string {
	return fmt.Sprintf("path=%q kept=%s dropped=%s", c.Path, c.Kept, c.Dropped)
}

type node struct {
	name string

	short        string
	long         string
	explicitDesc bool

	// group is true if the node has been declared as a group, or if it has children
	group       bool
	groupOrigin string

	children map[string]*node
	leaf     *Leaf
	mount    *Mount
}

func newNode(name string) *node {
	return &node{
		name:     name,
		short:    defaultGroupDescription(name),
		children: make(map[string]*node),
	}
}

func defaultGroupDescription(name string) string {
	return fmt.Sprintf("%s command group", name)
}

// Tree is the command tree of a single extension
type Tree struct {
	root *node

	// OnConflict is called for every command which is dropped due to a conflict
	OnConflict func(Conflict)
}

// New creates a new command tree. name is the root command name (i.e. the extension name)
func New(name string, short string) *Tree {
	root := newNode(name)
	root.short = short
	root.explicitDesc = true
	root.group = true
	return &Tree{root: root}
}

func (t *Tree) conflict(path []string, kept, dropped string) {
	if t.OnConflict != nil {
		t.OnConflict(Conflict{
			Path:    strings.Join(append([]string{t.root.name}, path...), " "),
			Kept:    kept,
			Dropped: dropped,
		})
	}
}

// ensureGroup returns the group node at the given path, creating any missing nodes.
// If a node on the path is currently a leaf, the leaf is dropped as groups take
// precedence over leaves.
func (t *Tree) ensureGroup(path []string, origin string) *node {
	current := t.root
	for i, name := range path {
		child, ok := current.children[name]
		if !ok {
			child = newNode(name)
			current.children[name] = child
		}
		if !child.group {
			child.group = true
			child.groupOrigin = origin
			if child.leaf != nil {
				t.conflict(path[:i+1], origin, child.leaf.Origin)
				child.leaf = nil
			}
		}
		current = child
	}
	return current
}

// AddGroup declares a command group at the given path. Groups are merged with
// any existing group at the same path. The first non-empty description is used,
// and it replaces the generated default description.
func (t *Tree) AddGroup(path []string, short, long, origin string) {
	if len(path) == 0 {
		return
	}
	group := t.ensureGroup(path, origin)
	if !group.explicitDesc && (short != "" || long != "") {
		group.short = short
		group.long = long
		group.explicitDesc = true
	}
}

// AddLeaf adds a runnable command at the given path. Missing parent groups are created.
//
// Conflicts are resolved as follows:
//   - if a group already exists at the path, the leaf is dropped
//   - if a leaf already exists at the path, the leaf with the higher precedence
//     (lower SourceKind) is kept. For equal precedence, the first leaf is kept.
func (t *Tree) AddLeaf(path []string, leaf Leaf) {
	if len(path) == 0 {
		return
	}
	parent := t.ensureGroup(path[:len(path)-1], leaf.Origin)
	name := path[len(path)-1]

	existing, ok := parent.children[name]
	if !ok {
		existing = newNode(name)
		existing.leaf = &leaf
		parent.children[name] = existing
		return
	}

	if existing.mount != nil && !existing.group {
		// A runnable command at the mount path takes precedence over the mount
		t.conflict(path, leaf.Origin, existing.mount.Origin)
		existing.mount = nil
		existing.leaf = &leaf
		return
	}

	if existing.group {
		t.conflict(path, existing.groupOrigin, leaf.Origin)
		return
	}

	if existing.leaf != nil && existing.leaf.Source <= leaf.Source {
		t.conflict(path, existing.leaf.Origin, leaf.Origin)
		return
	}

	if existing.leaf != nil {
		t.conflict(path, leaf.Origin, existing.leaf.Origin)
	}
	existing.leaf = &leaf
}

// AddMount adds a mount at the given path. An empty path mounts at the root of the tree.
// If a runnable command already exists at the path, the mount is dropped. If a group exists
// at the path, then the group's commands are used for their paths, and any other arguments
// are passed to the mount. If more than one mount is added at the same path, the first one is kept.
func (t *Tree) AddMount(path []string, mount Mount) {
	var target *node
	if len(path) == 0 {
		target = t.root
	} else {
		parent := t.ensureGroup(path[:len(path)-1], mount.Origin)
		name := path[len(path)-1]
		existing, ok := parent.children[name]
		if !ok {
			existing = newNode(name)
			parent.children[name] = existing
		}
		target = existing
	}

	if target.leaf != nil {
		t.conflict(path, target.leaf.Origin, mount.Origin)
		return
	}
	if target.mount != nil {
		t.conflict(path, target.mount.Origin, mount.Origin)
		return
	}
	target.mount = &mount
}

// IsEmpty returns true if the tree does not contain any runnable commands
func (t *Tree) IsEmpty() bool {
	return !hasLeaf(t.root)
}

func hasLeaf(n *node) bool {
	if n.leaf != nil || n.mount != nil {
		return true
	}
	for _, child := range n.children {
		if hasLeaf(child) {
			return true
		}
	}
	return false
}

// ToCobra converts the tree into cobra commands. Groups which do not contain any
// runnable commands are omitted. Nil is returned if the tree has no runnable commands.
// Errors from building individual commands are collected and returned together with
// the remaining commands.
func (t *Tree) ToCobra() (*cobra.Command, error) {
	var errs []error
	cmd := toCobra(t.root, &errs)
	if len(errs) > 0 {
		return cmd, fmt.Errorf("failed to build some commands: %w", errors.Join(errs...))
	}
	return cmd, nil
}

func toCobra(n *node, errs *[]error) *cobra.Command {
	if n.leaf != nil {
		cmd, err := n.leaf.Build()
		if err != nil {
			*errs = append(*errs, fmt.Errorf("%s: %w", n.leaf.Origin, err))
			return nil
		}
		return cmd
	}

	names := make([]string, 0, len(n.children))
	for name := range n.children {
		names = append(names, name)
	}
	sort.Strings(names)

	group := &cobra.Command{
		Use:   n.name,
		Short: n.short,
		Long:  n.long,
	}
	for _, name := range names {
		if child := toCobra(n.children[name], errs); child != nil {
			group.AddCommand(child)
		}
	}

	if n.mount != nil {
		if err := configureMount(group, n, n.mount); err != nil {
			*errs = append(*errs, fmt.Errorf("%s: %w", n.mount.Origin, err))
			if !group.HasSubCommands() {
				return nil
			}
		}
		return group
	}

	if !group.HasSubCommands() {
		return nil
	}
	return group
}

// IsHelpRequest returns true if the arguments only request help for the command
func IsHelpRequest(args []string) bool {
	return len(args) == 1 && (args[0] == "-h" || args[0] == "--help")
}

func configureMount(cmd *cobra.Command, n *node, mount *Mount) error {
	if !n.explicitDesc && mount.Short != "" {
		cmd.Short = mount.Short
	}
	if mount.Long != "" {
		if cmd.Long != "" {
			cmd.Long += "\n\n" + mount.Long
		} else {
			cmd.Long = mount.Long
		}
	}

	// Pass all arguments (including flags) to the mount
	cmd.Args = cobra.ArbitraryArgs
	cmd.DisableFlagParsing = true

	if err := mount.Configure(cmd); err != nil {
		return err
	}
	if cmd.RunE == nil && cmd.Run == nil {
		return fmt.Errorf("mount did not set a run function")
	}

	if cmd.HasSubCommands() {
		// Show the help of the group (which lists the other commands), as the
		// mount's own help would not include them
		run := cmd.RunE
		cmd.RunE = func(c *cobra.Command, args []string) error {
			if IsHelpRequest(args) {
				return c.Help()
			}
			return run(c, args)
		}
	}
	return nil
}
